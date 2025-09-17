package health

import (
	"context"
	"gophermart-service/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewHealthRepository(logger config.LoggerInterface, pool *pgxpool.Pool) RepositoryInterface {
	return &Repository{
		logger: logger,
		pool:   pool,
	}
}

type Repository struct {
	logger config.LoggerInterface
	pool   *pgxpool.Pool
}

func (r *Repository) Ping(ctx context.Context) error {
	query := `SELECT 1`
	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		return err
	}
	return nil
}
