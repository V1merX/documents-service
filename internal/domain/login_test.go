package domain_test

import (
	"testing"

	"github.com/V1merX/documents-service/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestNewLogin(t *testing.T) {
	t.Run("success Latin alphabet", func(t *testing.T) {
		wantLogin := "testuser"

		login, err := domain.NewLogin(wantLogin)
		require.NoError(t, err)

		require.Equal(t, wantLogin, login.Value())
	})

	t.Run("success with numbers", func(t *testing.T) {
		wantLogin := "testuser1234"

		login, err := domain.NewLogin(wantLogin)
		require.NoError(t, err)

		require.Equal(t, wantLogin, login.Value())
	})

	t.Run("error when login too short", func(t *testing.T) {
		login := "login12"

		_, err := domain.NewLogin(login)
		require.ErrorIs(t, err, domain.ErrLoginTooShort)
	})

	t.Run("error with non-ASCII characters", func(t *testing.T) {
		login := "иванпетров"

		_, err := domain.NewLogin(login)
		require.ErrorIs(t, err, domain.ErrInvalidLoginFormat)
	})

	t.Run("error with non-ASCII characters", func(t *testing.T) {
		login := "test_user"

		_, err := domain.NewLogin(login)
		require.ErrorIs(t, err, domain.ErrInvalidLoginFormat)
	})
}
