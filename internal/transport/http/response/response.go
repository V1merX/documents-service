package response

import (
	"encoding/json"
	"errors"
	"net/http"
)

func Data(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, Envelope{Data: data})
}

func Response(w http.ResponseWriter, status int, resp any) {
	writeJSON(w, status, Envelope{Response: resp})
}

func Fail(w http.ResponseWriter, err error, onInternal func()) {
	for sentinel, meta := range errorStatus {
		if errors.Is(err, sentinel) {
			writeJSON(w, meta.HTTPStatus, Envelope{Error: &Error{
				Code: meta.Code,
				Text: sentinel.Error(),
			}})
			return
		}
	}

	if onInternal != nil {
		onInternal()
	}
	writeJSON(w, http.StatusInternalServerError, Envelope{Error: &Error{
		Code: CodeInternal,
		Text: "internal server error",
	}})
}

func writeJSON(w http.ResponseWriter, status int, v Envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(v)
}
