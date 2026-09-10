package postgres

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"github.com/V1merX/documents-service/internal/domain"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Save(ctx context.Context, user *domain.User) error {
	_, err := r.pool.Exec(ctx, "INSERT INTO users(id, login, password) VALUES ($1, $2, $3)",
		user.ID(), user.Login().Value(), user.Password().Encoded())
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == pgerrcode.UniqueViolation {
			return domain.ErrUserAlreadyExists
		}
		return fmt.Errorf("save user: %w", err)
	}

	return nil
}

func (r *UserRepo) GetByLogin(ctx context.Context, login domain.Login) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, "SELECT id, password, token FROM users WHERE login = $1", login.Value())

	var (
		id       uuid.UUID
		password string
		token    *string
	)

	if err := row.Scan(&id, &password, &token); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by login: %w", err)
	}

	var t string
	if token != nil {
		t = *token
	}

	return domain.RestoreUser(id, login.Value(), password, t), nil
}

func (r *UserRepo) GetByToken(ctx context.Context, token domain.Token) (*domain.User, error) {
	row := r.pool.QueryRow(ctx, "SELECT id, login, password FROM users WHERE token = $1", token.Value())

	var (
		id              uuid.UUID
		login, password string
	)

	if err := row.Scan(&id, &login, &password); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by token: %w", err)
	}

	return domain.RestoreUser(id, login, password, token.Value()), nil
}

func (r *UserRepo) RevokeToken(ctx context.Context, token domain.Token) error {
	_, err := r.pool.Exec(ctx, "UPDATE users SET token = NULL WHERE token = $1", token.Value())
	if err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}

	return nil
}

func (r *UserRepo) UpdateToken(ctx context.Context, userID uuid.UUID, token domain.Token) error {
	tag, err := r.pool.Exec(ctx, "UPDATE users SET token = $1 WHERE id = $2", token.Value(), userID)
	if err != nil {
		return fmt.Errorf("update token: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}
