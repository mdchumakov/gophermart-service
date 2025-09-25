package withdraw

import (
	"context"
	"errors"
	"gophermart-service/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Специальная ошибка для недостаточного баланса
var ErrInsufficientBalance = errors.New("insufficient balance")

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

func (r Repository) AddNewWithBalanceCheck(ctx context.Context, userID int, orderNumber string, sum float32) error {
	// Начинаем транзакцию с уровнем изоляции SERIALIZABLE
	txOptions := pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	}
	tx, err := r.pool.BeginTx(ctx, txOptions)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Блокируем пользователя для чтения его баланса с FOR UPDATE
	// Это предотвращает другие транзакции от изменения баланса пользователя
	var currentBalance float32
	balanceQuery := `
		SELECT 
			COALESCE(SUM(CASE WHEN o.status = 'PROCESSED' THEN o.accrual ELSE 0 END), 0) - 
			COALESCE(SUM(w.sum), 0) as current_balance
		FROM users u
		LEFT JOIN orders o ON u.id = o.user_id
		LEFT JOIN withdrawals w ON u.id = w.user_id
		WHERE u.id = $1
		GROUP BY u.id
		FOR UPDATE OF u`

	err = tx.QueryRow(ctx, balanceQuery, userID).Scan(&currentBalance)
	if err != nil {
		return err
	}

	// Проверяем, достаточно ли средств
	if currentBalance < sum {
		return ErrInsufficientBalance
	}

	// Добавляем списание
	insertQuery := `INSERT INTO withdrawals (user_id, order_number, sum) VALUES ($1, $2, $3)`
	_, err = tx.Exec(ctx, insertQuery, userID, orderNumber, sum)
	if err != nil {
		return err
	}

	// Коммитим транзакцию
	if err = tx.Commit(ctx); err != nil {
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
