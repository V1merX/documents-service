package response

import (
	"net/http"
	"time"

	"github.com/V1merX/documents-service/internal/domain"
)

const (
	CodeBadRequest     = 100
	CodeValidation     = 101
	CodeUnauthorized   = 200
	CodeBadCredentials = 201
	CodeForbidden      = 300
	CodeNotFound       = 400
	CodeInternal       = 500
	CodeNotImplemented = 501
)

type errorMeta struct {
	HTTPStatus int
	Code       int
}

var errorStatus = map[error]errorMeta{
	domain.ErrValueOrKeyEmpty:        {http.StatusBadRequest, CodeValidation},
	domain.ErrNonIntegerLimit:        {http.StatusBadRequest, CodeValidation},
	domain.ErrInvalidLimit:           {http.StatusBadRequest, CodeValidation},
	domain.ErrLoginTooShort:          {http.StatusBadRequest, CodeValidation},
	domain.ErrInvalidLoginFormat:     {http.StatusBadRequest, CodeValidation},
	domain.ErrPasswordTooShort:       {http.StatusBadRequest, CodeValidation},
	domain.ErrPasswordMissingCase:    {http.StatusBadRequest, CodeValidation},
	domain.ErrPasswordMissingDigit:   {http.StatusBadRequest, CodeValidation},
	domain.ErrPasswordMissingSpecial: {http.StatusBadRequest, CodeValidation},
	domain.ErrInvalidPasswordFormat:  {http.StatusBadRequest, CodeValidation},
	domain.ErrEmptyFileName:          {http.StatusBadRequest, CodeValidation},
	domain.ErrInvalidOwner:           {http.StatusBadRequest, CodeValidation},
	domain.ErrRequiredMime:           {http.StatusBadRequest, CodeValidation},
	domain.ErrInvalidDocumentID:      {http.StatusBadRequest, CodeValidation},
	domain.ErrInvalidFilterKey:       {http.StatusBadRequest, CodeValidation},
	domain.ErrUserAlreadyExists:      {http.StatusBadRequest, CodeValidation},

	domain.ErrBadRequest: {http.StatusBadRequest, CodeBadRequest},

	domain.ErrEmptyToken:        {http.StatusUnauthorized, CodeUnauthorized},
	domain.ErrInvalidToken:      {http.StatusUnauthorized, CodeUnauthorized},
	domain.ErrIncorrectPassword: {http.StatusUnauthorized, CodeBadCredentials},
	domain.ErrUserNotFound:      {http.StatusUnauthorized, CodeBadCredentials},

	domain.ErrForbidden: {http.StatusForbidden, CodeForbidden},

	domain.ErrNotFound:         {http.StatusNotFound, CodeNotFound},
	domain.ErrDocumentNotFound: {http.StatusNotFound, CodeNotFound},
}

type Envelope struct {
	Error    *Error `json:"error,omitempty"`
	Response any    `json:"response,omitempty"`
	Data     any    `json:"data,omitempty"`
}

type Error struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

type RegisterResponse struct {
	Login string `json:"login"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type CreateDocsResponse struct {
	JSON any    `json:"json,omitempty"`
	File string `json:"file,omitempty"`
}

type ListDocsResponse struct {
	Docs []Doc `json:"docs"`
}

type Doc struct {
	ID      string         `json:"id"`
	Name    string         `json:"name"`
	Mime    string         `json:"mime"`
	File    bool           `json:"file"`
	Public  bool           `json:"public"`
	Created time.Time      `json:"created"`
	Grant   []domain.Login `json:"grant"`
}
