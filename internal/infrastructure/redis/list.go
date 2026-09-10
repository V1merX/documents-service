package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/V1merX/documents-service/internal/domain"
	"github.com/V1merX/documents-service/internal/infrastructure/redis/convertor"
	"github.com/V1merX/documents-service/internal/infrastructure/redis/model"
	"github.com/V1merX/documents-service/internal/service/document"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	listCacheKey  = "lists:"
	listIndexKey  = "lists:index:"
	listIndexHold = time.Hour
)

type ListCache struct {
	client *redis.Client
	log    *zap.Logger
	ttl    time.Duration
}

func NewListCache(client *redis.Client, log *zap.Logger, ttl time.Duration) *ListCache {
	return &ListCache{client: client, log: log, ttl: ttl}
}

func (c *ListCache) Get(ctx context.Context, key document.ListKey) (*[]domain.Document, bool, error) {
	cacheKey := listKey(key)

	data, err := c.client.Get(ctx, cacheKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}

		return nil, false, fmt.Errorf("get list from cache: %w", err)
	}

	var cached []model.Document
	if err := json.Unmarshal(data, &cached); err != nil {
		c.log.Error("Failed to unmarshal cached list",
			zap.Error(err), zap.String("key", cacheKey))

		if delErr := c.client.Del(ctx, cacheKey).Err(); delErr != nil {
			c.log.Error("Failed to drop corrupted list entry",
				zap.Error(delErr), zap.String("key", cacheKey))
		}

		return nil, false, nil
	}

	docs := convertor.DocsToDomain(cached)

	return &docs, true, nil
}

func (c *ListCache) Put(ctx context.Context, key document.ListKey, docs *[]domain.Document) error {
	if docs == nil {
		return nil
	}

	data, err := json.Marshal(convertor.DocsToModel(*docs))
	if err != nil {
		return fmt.Errorf("marshal list for cache: %w", err)
	}

	cacheKey := listKey(key)

	pipe := c.client.TxPipeline()
	pipe.Set(ctx, cacheKey, data, c.ttl)

	for _, login := range indexLogins(key) {
		index := listIndexKey + login
		pipe.SAdd(ctx, index, cacheKey)
		pipe.Expire(ctx, index, c.ttl+listIndexHold)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("put list in cache: %w", err)
	}

	return nil
}

func (c *ListCache) Invalidate(ctx context.Context, logins ...domain.Login) error {
	indexes := make([]string, 0, len(logins))
	seen := make(map[string]struct{}, len(logins))

	for _, login := range logins {
		value := login.Value()
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}

		seen[value] = struct{}{}
		indexes = append(indexes, listIndexKey+value)
	}

	if len(indexes) == 0 {
		return nil
	}

	members, err := c.client.SUnion(ctx, indexes...).Result()
	if err != nil {
		return fmt.Errorf("collect list keys for invalidation: %w", err)
	}

	pipe := c.client.TxPipeline()
	if len(members) > 0 {
		pipe.Del(ctx, members...)
	}
	pipe.Del(ctx, indexes...)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("invalidate lists: %w", err)
	}

	return nil
}

func listKey(key document.ListKey) string {
	var b strings.Builder

	b.WriteString(key.Requester.Value())
	b.WriteByte('|')
	b.WriteString(key.Target.Value())
	b.WriteByte('|')
	b.WriteString(deref(key.Key))
	b.WriteByte('|')
	b.WriteString(deref(key.Value))
	b.WriteByte('|')
	b.WriteString(strconv.Itoa(key.Limit))

	sum := sha256.Sum256([]byte(b.String()))

	return listCacheKey + hex.EncodeToString(sum[:])
}

func indexLogins(key document.ListKey) []string {
	logins := make([]string, 0, 2)

	if requester := key.Requester.Value(); requester != "" {
		logins = append(logins, requester)
	}

	if target := key.Target.Value(); target != "" && target != key.Requester.Value() {
		logins = append(logins, target)
	}

	return logins
}

func deref(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
