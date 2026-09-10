package domain

import (
	"encoding/json"
	"regexp"
	"unicode/utf8"
)

const (
	minLoginLength = 8
)

var loginRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

type Login struct {
	value string
}

func NewLogin(input string) (Login, error) {
	if err := ValidateLogin(input); err != nil {
		return Login{}, err
	}

	return Login{value: input}, nil
}

func ValidateLogin(login string) error {
	if utf8.RuneCountInString(login) < minLoginLength {
		return ErrLoginTooShort
	}

	if !loginRegex.MatchString(login) {
		return ErrInvalidLoginFormat
	}

	return nil
}

func (l Login) Value() string {
	return l.value
}

func (l Login) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.value)
}

func RestoreLogin(input string) Login {
	return Login{value: input}
}
