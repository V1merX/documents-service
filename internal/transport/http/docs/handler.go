package docs

import (
	"context"

	"github.com/V1merX/documents-service/internal/domain"
	"github.com/V1merX/documents-service/internal/service/document"
	"go.uber.org/zap"
)

type service interface {
	Create(ctx context.Context, cmd document.CreateCommand) (*domain.Document, error)
	List(ctx context.Context, cmd document.ListCommand) (*[]domain.Document, error)
	Get(ctx context.Context, cmd document.GetCommand) (*domain.Document, error)
	Delete(ctx context.Context, cmd document.DeleteCommand) (string, error)
}

type userService interface {
	Exists(ctx context.Context, token string) (domain.Login, error)
}

type Handler struct {
	log         *zap.Logger
	svc         service
	userService userService
}

func NewHandler(log *zap.Logger, svc service, userService userService) *Handler {
	return &Handler{
		log:         log,
		svc:         svc,
		userService: userService,
	}
}
