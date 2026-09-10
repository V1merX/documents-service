package auth

import (
	"net/http"

	"github.com/V1merX/documents-service/internal/transport/http/convertor"
	"github.com/V1merX/documents-service/internal/transport/http/request"
	"github.com/V1merX/documents-service/internal/transport/http/response"
	"go.uber.org/zap"
)

func (h *Handler) Auth(w http.ResponseWriter, r *http.Request) {
	req := request.AuthRequest{
		Login:    r.PostFormValue("login"),
		Password: r.PostFormValue("pswd"),
	}

	token, err := h.svc.Auth(r.Context(), convertor.ToAuthCommand(req))
	if err != nil {
		response.Fail(w, err, func() {
			h.log.Error("Failed to authenticate", zap.Error(err), zap.String("login", req.Login))
		})
		return
	}

	response.Response(w, http.StatusOK, convertor.ToAuthResponse(token.Value()))
}
