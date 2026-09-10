package domain

import (
	"crypto/rand"
)

type Token struct {
	value string
}

func GenerateToken() Token {
	return Token{
		value: rand.Text(),
	}
}

func ValidateToken(token string) (Token, error) {
	if len(token) == 0 {
		return Token{}, ErrEmptyToken
	}

	return Token{
		value: token,
	}, nil
}

func (t *Token) Value() string {
	return t.value
}

func RestoreToken(input string) Token {
	return Token{
		value: input,
	}
}
