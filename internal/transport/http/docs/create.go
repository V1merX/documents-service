package docs

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/V1merX/documents-service/internal/domain"
	"github.com/V1merX/documents-service/internal/transport/http/convertor"
	"github.com/V1merX/documents-service/internal/transport/http/middleware"
	"github.com/V1merX/documents-service/internal/transport/http/request"
	"github.com/V1merX/documents-service/internal/transport/http/response"
	"go.uber.org/zap"
)

const (
	maxUploadSize   int64 = 32 << 20
	maxMultipartMem int64 = 8 << 20
)

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	login, exists := middleware.UserFromContext(r.Context())
	if !exists {
		response.Fail(w, domain.ErrEmptyToken, nil)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxMultipartMem); err != nil {
		response.Fail(w, domain.ErrBadRequest, nil)
		return
	}
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()

	var meta request.DocumentMeta
	if err := json.Unmarshal([]byte(r.FormValue("meta")), &meta); err != nil {
		response.Fail(w, domain.ErrBadRequest, nil)
		return
	}

	if meta.Token == "" {
		response.Fail(w, domain.ErrEmptyToken, nil)
		return
	}

	var jsonData []byte
	if raw := r.FormValue("json"); raw != "" {
		if !json.Valid([]byte(raw)) {
			response.Fail(w, domain.ErrBadRequest, nil)
			return
		}

		jsonData = []byte(raw)
	}

	var content []byte
	if meta.File {
		file, header, err := r.FormFile("file")
		if err != nil {
			response.Fail(w, domain.ErrBadRequest, nil)
			return
		}
		defer file.Close()

		content, err = io.ReadAll(file)
		if err != nil {
			response.Fail(w, err, func() {
				h.log.Error("Failed to read uploaded file", zap.Error(err), zap.String("name", meta.Name))
			})
			return
		}

		if meta.Name == "" {
			meta.Name = header.Filename
		}
	}

	if meta.Name == "" {
		response.Fail(w, domain.ErrEmptyFileName, nil)
		return
	}

	if !meta.File && len(jsonData) == 0 {
		response.Fail(w, domain.ErrBadRequest, nil)
		return
	}

	if _, err := h.svc.Create(r.Context(), convertor.ToDocCreateCommand(meta, login, jsonData, content)); err != nil {
		response.Fail(w, err, func() {
			h.log.Error("Failed to create document", zap.Error(err), zap.String("name", meta.Name))
		})
		return
	}

	var fileName string
	if meta.File {
		fileName = meta.Name
	}

	response.Data(w, http.StatusOK, convertor.ToCreateDocsResponse(fileName, jsonData))
}
