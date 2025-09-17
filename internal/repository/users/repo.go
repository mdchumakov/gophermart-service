package users

import (
	"context"
	"errors"
	"gophermart-service/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewUsersRepository(logger config.LoggerInterface, pool *pgxpool.Pool) RepositoryInterface {
	return &Repository{
		logger: logger,
		pool:   pool,
	}
}

type Repository struct {
	logger config.LoggerInterface
	pool   *pgxpool.Pool
}

func (r *Repository) Add(ctx context.Context, username, passwordHash string) (int, error) {
	query := `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id`

	var userID int
	err := r.pool.QueryRow(ctx, query, username, passwordHash).Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" || err.Error() == "UNIQUE constraint failed" {
				return 0, ErrUserLoginAlreadyExists
			}
		}
		return 0, err
	}
	return userID, nil
}

func (r *Repository) GetUserHashPassword(ctx context.Context, login string) (int, string, error) {
	query := `SELECT id, password_hash FROM users WHERE login = $1`

	var (
		userID       int
		passwordHash string
	)
	err := r.pool.QueryRow(ctx, query, login).Scan(&userID, &passwordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, "", ErrUserNotFound
		}
		return 0, "", err
	}
	return userID, passwordHash, nil
}
