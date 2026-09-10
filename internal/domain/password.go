package domain

import (
	"regexp"
	"unicode"
	"unicode/utf8"
)

const (
	minPasswordLength = 8
)

var passwordRegex = regexp.MustCompile(`\A[a-zA-Z0-9а-яА-ЯёЁ!@#$%^&*()_+\-=\[\]{};':"\\|,.<>?]+\z`)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, encodedHash string) (bool, error)
}

type Password struct {
	encoded string
}

func NewPassword(input string, hasher PasswordHasher) (Password, error) {
	if utf8.RuneCountInString(input) < minPasswordLength {
		return Password{}, ErrPasswordTooShort
	}

	var (
		hasLower   bool
		hasUpper   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, r := range []rune(input) {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		case !unicode.IsLetter(r) && !unicode.IsDigit(r):
			hasSpecial = true
		}
	}

	if !hasLower || !hasUpper {
		return Password{}, ErrPasswordMissingCase
	}
	if !hasDigit {
		return Password{}, ErrPasswordMissingDigit
	}
	if !hasSpecial {
		return Password{}, ErrPasswordMissingSpecial
	}

	if !passwordRegex.MatchString(input) {
		return Password{}, ErrInvalidPasswordFormat
	}

	hashedPass, err := hasher.Hash(input)
	if err != nil {
		return Password{}, err
	}

	return Password{
		encoded: hashedPass,
	}, nil
}

func (p Password) Encoded() string {
	return p.encoded
}

func (p Password) Verify(password string, hasher PasswordHasher) error {
	ok, err := hasher.Verify(password, p.encoded)
	if err != nil {
		return err
	}
	if !ok {
		return ErrIncorrectPassword
	}

	return nil
}

func RestorePassword(encoded string) Password {
	return Password{
		encoded: encoded,
	}
}
