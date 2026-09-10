package docs

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/V1merX/documents-service/internal/domain"
	"github.com/V1merX/documents-service/internal/transport/http/convertor"
	"github.com/V1merX/documents-service/internal/transport/http/middleware"
	"github.com/V1merX/documents-service/internal/transport/http/response"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	login, exists := middleware.UserFromContext(r.Context())
	if !exists {
		response.Fail(w, domain.ErrEmptyToken, nil)
		return
	}

	doc, err := h.svc.Get(r.Context(), convertor.ToDocGetCommand(login, id))
	if err != nil {
		response.Fail(w, err, func() {
			h.log.Error("Failed to get document", zap.Error(err), zap.String("id", id))
		})
		return
	}

	if doc.File() {
		if mime := doc.Mime(); mime != "" {
			w.Header().Set("Content-Type", mime)
		}

		http.ServeContent(w, r, doc.Name(), doc.Created(), bytes.NewReader(doc.Content()))

		return
	}

	data := doc.JSON()
	if len(data) == 0 {
		data = []byte("{}")
	}

	response.Data(w, http.StatusOK, json.RawMessage(data))
}
