package docs

import (
	"net/http"
	"strconv"

	"github.com/V1merX/documents-service/internal/domain"
	"github.com/V1merX/documents-service/internal/transport/http/convertor"
	"github.com/V1merX/documents-service/internal/transport/http/middleware"
	"github.com/V1merX/documents-service/internal/transport/http/response"
	"go.uber.org/zap"
)

const (
	defaultLimit int = 20
	maxLimit     int = 100
)

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	requesterLogin, exists := middleware.UserFromContext(r.Context())
	if !exists {
		response.Fail(w, domain.ErrEmptyToken, nil)
		return
	}

	q := r.URL.Query()

	login := q.Get("login")
	key := q.Get("key")
	value := q.Get("value")
	limitStr := q.Get("limit")

	if (key != "" && value == "") || (key == "" && value != "") {
		response.Fail(w, domain.ErrValueOrKeyEmpty, nil)
		return
	}

	limit := defaultLimit
	if limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err != nil {
			response.Fail(w, domain.ErrNonIntegerLimit, nil)
			return
		}
		limit = parsed
	}
	if limit <= 0 || limit > maxLimit {
		response.Fail(w, domain.ErrInvalidLimit, nil)
		return
	}

	docs, err := h.svc.List(r.Context(), convertor.ToDocListCommand(requesterLogin, limit, key, value, login))
	if err != nil {
		response.Fail(w, err, func() {
			h.log.Error("Failed to get documents",
				zap.Error(err),
				zap.String("login", login),
				zap.String("key", key),
				zap.String("value", value),
				zap.Int("limit", limit),
			)
		})
		return
	}

	response.Data(w, http.StatusOK, convertor.ToDocumentsResponse(docs))
}
