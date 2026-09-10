package middleware

import (
	"context"
	"net/http"
	"uuid"

	"github.com/V1merX/documents-service/internal/domain"
	"github.com/V1merX/documents-service/internal/transport/http/response"
	"go.uber.org/zap"
)

type TokenChecker interface {
	Exists(ctx context.Context, token string) (uuid.UUID, error)
}

type tokenCtxKey struct{}

func AuthMiddleware(storer TokenChecker, log *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.URL.Query().Get("token")
			if token == "" {
				response.Fail(w, domain.ErrEmptyToken, nil)
				return
			}

			userID, err := storer.Exists(r.Context(), token)
			if err != nil {
				response.Fail(w, err, func() {
					log.Error("Failed to check token", zap.Error(err))
				})
				return
			}

			if userID == uuid.Nil() {
				response.Fail(w, domain.ErrInvalidToken, nil)
				return
			}

			ctx := context.WithValue(r.Context(), tokenCtxKey{}, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserFromContext(ctx context.Context) (domain.Login, bool) {
	login, ok := ctx.Value(tokenCtxKey{}).(domain.Login)
	return login, ok
}
