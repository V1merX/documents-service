package docs

import (
	"net/http"

	"github.com/V1merX/documents-service/internal/domain"
	"github.com/V1merX/documents-service/internal/transport/http/convertor"
	"github.com/V1merX/documents-service/internal/transport/http/middleware"
	"github.com/V1merX/documents-service/internal/transport/http/response"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	login, exists := middleware.UserFromContext(r.Context())
	if !exists {
		response.Fail(w, domain.ErrEmptyToken, nil)
		return
	}

	id := chi.URLParam(r, "id")

	deletedID, err := h.svc.Delete(r.Context(), convertor.ToDocDeleteCommand(login, id))
	if err != nil {
		response.Fail(w, err, func() {
			h.log.Error("Failed to delete document", zap.Error(err), zap.String("id", id))
		})
		return
	}

	response.Response(w, http.StatusOK, convertor.ToDocDeleteResponse(deletedID))
}
