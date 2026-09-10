package domain

import "errors"

var (
	ErrBadRequest = errors.New("bad request")
	ErrNotFound   = errors.New("not found")
	ErrForbidden  = errors.New("forbidden")

	ErrValueOrKeyEmpty = errors.New("value or key empty")
	ErrNonIntegerLimit = errors.New("limit must be an integer")
	ErrInvalidLimit    = errors.New("limit must be between 1 and 100")

	ErrLoginTooShort      = errors.New("login too short")
	ErrInvalidLoginFormat = errors.New("invalid login format")

	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")

	ErrPasswordTooShort       = errors.New("password too short")
	ErrPasswordMissingCase    = errors.New("password missing case")
	ErrPasswordMissingDigit   = errors.New("password missing digit")
	ErrPasswordMissingSpecial = errors.New("password missing special")
	ErrInvalidPasswordFormat  = errors.New("invalid password format")
	ErrIncorrectPassword      = errors.New("incorrect password")

	ErrEmptyToken   = errors.New("empty token")
	ErrInvalidToken = errors.New("invalid token")

	ErrEmptyFileName = errors.New("empty filename")
	ErrInvalidOwner  = errors.New("invalid owner")
	ErrRequiredMime  = errors.New("mime is required")

	ErrInvalidDocumentID = errors.New("invalid document ID")
	ErrDocumentNotFound  = errors.New("document not found")
	ErrInvalidFilterKey  = errors.New("invalid filter key")
)
