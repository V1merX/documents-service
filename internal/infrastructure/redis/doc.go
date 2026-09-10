package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/V1merX/documents-service/internal/domain"
	"github.com/V1merX/documents-service/internal/infrastructure/redis/convertor"
	"github.com/V1merX/documents-service/internal/infrastructure/redis/model"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	documentCacheKey    = "documents:"
	maxCacheableContent = 1 << 20 // 1 MiB
)

type DocumentCache struct {
	client *redis.Client
	log    *zap.Logger
	ttl    time.Duration
}

func NewDocumentCache(client *redis.Client, logger *zap.Logger, ttl time.Duration) *DocumentCache {
	return &DocumentCache{client: client, log: logger, ttl: ttl}
}

func (c *DocumentCache) Get(ctx context.Context, id uuid.UUID) (*domain.Document, bool, error) {
	data, err := c.client.Get(ctx, documentCacheKey+id.String()).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}

		return nil, false, fmt.Errorf("get document from cache: %w", err)
	}

	var document model.Document
	if err := json.Unmarshal(data, &document); err != nil {
		c.log.Error("Failed to unmarshal cached document",
			zap.Error(err), zap.String("id", id.String()))

		if delErr := c.Delete(ctx, id); delErr != nil {
			c.log.Error("Failed to drop corrupted cache entry",
				zap.Error(delErr), zap.String("id", id.String()))
		}

		return nil, false, nil
	}

	return convertor.DocToDomain(document), true, nil
}

func (c *DocumentCache) Put(ctx context.Context, doc *domain.Document) error {
	if len(doc.Content()) > maxCacheableContent {
		return nil
	}

	data, err := json.Marshal(convertor.DocToModel(doc))
	if err != nil {
		return fmt.Errorf("marshal document for cache: %w", err)
	}

	if err := c.client.Set(ctx, documentCacheKey+doc.ID().String(), data, c.ttl).Err(); err != nil {
		return fmt.Errorf("put document in cache: %w", err)
	}

	return nil
}

func (c *DocumentCache) Delete(ctx context.Context, id uuid.UUID) error {
	if err := c.client.Del(ctx, documentCacheKey+id.String()).Err(); err != nil {
		return fmt.Errorf("delete document from cache: %w", err)
	}

	return nil
}
