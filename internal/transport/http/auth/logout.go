package auth

import (
	"net/http"

	"github.com/V1merX/documents-service/internal/transport/http/convertor"
	"github.com/V1merX/documents-service/internal/transport/http/response"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	tokenPath := chi.URLParam(r, "token")

	token, err := h.svc.Logout(r.Context(), convertor.ToLogoutCommand(tokenPath))
	if err != nil {
		response.Fail(w, err, func() {
			h.log.Error("Failed to logout", zap.Error(err))
		})
		return
	}

	response.Response(w, http.StatusOK, convertor.ToLogoutResponse(token.Value()))
}
