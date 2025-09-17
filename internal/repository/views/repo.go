package views

import (
	"context"
	"gophermart-service/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewViewsRepository(logger config.LoggerInterface, pool *pgxpool.Pool) RepositoryInterface {
	return &Repository{
		logger: logger,
		pool:   pool,
	}
}

type Repository struct {
	logger config.LoggerInterface
	pool   *pgxpool.Pool
}

func (r Repository) GetUserBalance(ctx context.Context, userID int) (*UserBalance, error) {
	query := `SELECT * FROM user_balance WHERE user_id = $1`

	var balance UserBalance
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&balance.UserID,
		&balance.TotalAccrued,
		&balance.TotalWithdrawn,
		&balance.CurrentBalance,
	)
	if err != nil {
		return nil, err
	}

	return &balance, nil
}
