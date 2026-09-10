package di

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"syscall"

	"github.com/V1merX/documents-service/internal/config"
	"github.com/V1merX/documents-service/internal/infrastructure/argon2"
	"github.com/V1merX/documents-service/internal/infrastructure/postgres"
	cache "github.com/V1merX/documents-service/internal/infrastructure/redis"
	"github.com/V1merX/documents-service/internal/infrastructure/storage"
	"github.com/V1merX/documents-service/internal/service/document"
	"github.com/V1merX/documents-service/internal/service/user"
	authH "github.com/V1merX/documents-service/internal/transport/http/auth"
	docsH "github.com/V1merX/documents-service/internal/transport/http/docs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Container struct {
	flushOnce  *sync.Once
	finalizers []func()

	logger      *zap.Logger
	conf        *config.Config
	pgxPool     *pgxpool.Pool
	redisClient *redis.Client

	authHandler *authH.Handler
	docsHandler *docsH.Handler

	userSvc *user.Service
	docSvc  *document.Service

	userRepo *postgres.UserRepo
	docRepo  *postgres.DocumentRepo

	fileStorage *storage.FileStorage

	listCache *cache.ListCache
	docCache  *cache.DocumentCache
}

func NewContainer() (*Container, error) {
	var (
		c   Container
		err error
	)

	c.flushOnce = new(sync.Once)

	c.conf = config.MustLoad()

	c.logger, err = zap.NewProduction()
	if err != nil {
		return nil, err
	}
	c.finalizers = append(c.finalizers, func() {
		if err := c.logger.Sync(); err != nil {
			if !errors.Is(err, syscall.EINVAL) && !errors.Is(err, syscall.ENOTTY) {
				slog.Error("Failed to sync logger", slog.Any("error", err))
			}
		}
	})

	c.pgxPool, err = pgxpool.New(context.Background(), c.Config().DatabaseURL)
	if err != nil {
		return nil, err
	}
	c.finalizers = append(c.finalizers, func() {
		c.pgxPool.Close()
	})

	c.redisClient = redis.NewClient(&redis.Options{
		Addr:     c.Config().Redis.Addr,
		Password: c.Config().Redis.Password,
		DB:       c.Config().Redis.DB,
	})
	c.finalizers = append(c.finalizers, func() {
		if err := c.redisClient.Close(); err != nil {
			slog.Error("Failed to close redis client", slog.Any("error", err))
		}
	})

	c.userRepo = postgres.NewUserRepo(c.pgxPool)
	c.fileStorage, err = storage.New(c.Config().StoragePath)
	if err != nil {
		return nil, err
	}

	c.docRepo = postgres.NewDocumentRepo(c.pgxPool, c.fileStorage, c.Logger())

	c.listCache = cache.NewListCache(c.redisClient, c.Logger(), c.Config().Redis.TTL)
	c.docCache = cache.NewDocumentCache(c.redisClient, c.Logger(), c.Config().Redis.TTL)

	c.userSvc = user.NewService([]byte(c.Config().AdminToken), new(argon2.NewDefault()), c.userRepo)
	c.docSvc = document.NewService(c.Logger(), c.docRepo, c.docCache, c.listCache)

	c.authHandler = authH.NewHandler(c.Logger(), c.UserService())
	c.docsHandler = docsH.NewHandler(c.Logger(), c.docSvc)

	return &c, err
}

func (c *Container) Flush() {
	c.flushOnce.Do(func() {
		for _, f := range c.finalizers {
			f()
		}
	})
}

func (c *Container) Config() *config.Config {
	return c.conf
}

func (c *Container) Logger() *zap.Logger {
	return c.logger
}

func (c *Container) AuthHandler() *authH.Handler {
	return c.authHandler
}

func (c *Container) DocsHandler() *docsH.Handler {
	return c.docsHandler
}

func (c *Container) UserService() *user.Service {
	return c.userSvc
}
