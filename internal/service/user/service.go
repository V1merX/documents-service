package user

import (
	"context"
	"crypto/subtle"
	"uuid"

	"github.com/V1merX/documents-service/internal/domain"
)

type repo interface {
	Save(ctx context.Context, user *domain.User) error
	GetByLogin(ctx context.Context, login domain.Login) (*domain.User, error)
	GetByToken(ctx context.Context, token domain.Token) (*domain.User, error)
	RevokeToken(ctx context.Context, token domain.Token) error
	UpdateToken(ctx context.Context, userID uuid.UUID, token domain.Token) error
}

type Service struct {
	adminToken []byte
	repo       repo
	hasher     domain.PasswordHasher
}

func NewService(adminToken []byte, hasher domain.PasswordHasher, repo repo) *Service {
	return &Service{
		adminToken: adminToken,
		hasher:     hasher,
		repo:       repo,
	}
}

func (s *Service) Exists(ctx context.Context, inputToken string) (domain.Login, error) {
	token, err := domain.ValidateToken(inputToken)
	if err != nil {
		return domain.Login{}, err
	}

	user, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return domain.Login{}, err
	}

	return user.Login(), nil
}

func (s *Service) Register(ctx context.Context, cmd RegisterCommand) (domain.Login, error) {
	if subtle.ConstantTimeCompare([]byte(cmd.Token), s.adminToken) != 1 {
		return domain.Login{}, domain.ErrInvalidToken
	}

	login, err := domain.NewLogin(cmd.Login)
	if err != nil {
		return domain.Login{}, err
	}

	password, err := domain.NewPassword(cmd.Password, s.hasher)
	if err != nil {
		return domain.Login{}, err
	}

	user := domain.NewUser(login, password)

	if err := s.repo.Save(ctx, user); err != nil {
		return domain.Login{}, err
	}

	return login, nil
}

func (s *Service) Auth(ctx context.Context, cmd AuthCommand) (domain.Token, error) {
	login, err := domain.NewLogin(cmd.Login)
	if err != nil {
		return domain.Token{}, err
	}

	user, err := s.repo.GetByLogin(ctx, login)
	if err != nil {
		return domain.Token{}, err
	}

	if err := user.Password().Verify(cmd.Password, s.hasher); err != nil {
		return domain.Token{}, err
	}

	token := domain.GenerateToken()

	if err := s.repo.UpdateToken(ctx, user.ID(), token); err != nil {
		return domain.Token{}, err
	}

	return token, nil
}

func (s *Service) Logout(ctx context.Context, cmd LogoutCommand) (domain.Token, error) {
	token, err := domain.ValidateToken(cmd.Token)
	if err != nil {
		return domain.Token{}, err
	}

	if err := s.repo.RevokeToken(ctx, token); err != nil {
		return domain.Token{}, err
	}

	return token, nil
}
