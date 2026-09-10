package auth

import (
	"context"

	"github.com/V1merX/documents-service/internal/domain"
	"github.com/V1merX/documents-service/internal/service/user"
	"go.uber.org/zap"
)

type service interface {
	Register(ctx context.Context, cmd user.RegisterCommand) (domain.Login, error)
	Auth(ctx context.Context, cmd user.AuthCommand) (domain.Token, error)
	Logout(ctx context.Context, cmd user.LogoutCommand) (domain.Token, error)
}

type Handler struct {
	log *zap.Logger
	svc service
}

func NewHandler(log *zap.Logger, svc service) *Handler {
	return &Handler{
		log: log,
		svc: svc,
	}
}
