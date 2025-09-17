package withdraw

import (
	"context"
	"gophermart-service/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewWithdrawRepository(logger config.LoggerInterface, pool *pgxpool.Pool) RepositoryInterface {
	return &Repository{
		logger: logger,
		pool:   pool,
	}
}

type Repository struct {
	logger config.LoggerInterface
	pool   *pgxpool.Pool
}

func (r Repository) AddNew(ctx context.Context, userID int, orderNumber string, sum float32) error {
	query := `INSERT INTO withdrawals (user_id, order_number, sum) VALUES ($1, $2, $3)`

	_, err := r.pool.Exec(ctx, query, userID, orderNumber, sum)
	if err != nil {
		return err
	}
	return nil
}

func (r Repository) GetUserWithdrawals(ctx context.Context, userID int) ([]Withdrawal, error) {
	query := `SELECT order_number, sum, processed_at FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []Withdrawal
	for rows.Next() {
		var withdrawal Withdrawal
		err := rows.Scan(&withdrawal.Order, &withdrawal.Sum, &withdrawal.ProcessedAt)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, withdrawal)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return withdrawals, nil
}
