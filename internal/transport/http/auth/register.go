package auth

import (
	"encoding/json"
	"net/http"

	"github.com/V1merX/documents-service/internal/domain"
	"github.com/V1merX/documents-service/internal/transport/http/convertor"
	"github.com/V1merX/documents-service/internal/transport/http/request"
	"github.com/V1merX/documents-service/internal/transport/http/response"
	"go.uber.org/zap"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req request.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Fail(w, domain.ErrBadRequest, nil)
		return
	}

	login, err := h.svc.Register(r.Context(), convertor.ToRegisterCommand(req))
	if err != nil {
		response.Fail(w, err, func() {
			h.log.Error("Failed to register user", zap.Error(err), zap.String("login", req.Login))
		})
		return
	}

	response.Response(w, http.StatusOK, convertor.ToRegisterResponse(login.Value()))
}
