package orders

import (
	"context"
	"errors"
	"gophermart-service/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Специальные ошибки для транзакционного добавления заказов
var (
	ErrOrderAlreadyExistsForUserRepo = errors.New("order already exists for this user")
	ErrOrderAlreadyProcessedRepo     = errors.New("order already processed by another user")
)

func NewOrdersRepository(logger config.LoggerInterface, pool *pgxpool.Pool) RepositoryInterface {
	return &Repository{
		logger: logger,
		pool:   pool,
	}
}

type Repository struct {
	logger config.LoggerInterface
	pool   *pgxpool.Pool
}

func (r *Repository) CheckUsersOrderExists(ctx context.Context, userID int, orderNumber string) (bool, error) {
	query := `SELECT 1 FROM orders WHERE user_id = $1 AND order_number = $2`
	var exists int
	err := r.pool.QueryRow(ctx, query, userID, orderNumber).Scan(&exists)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *Repository) CheckOrderAlreadyProcessed(ctx context.Context, userID int, orderNumber string) (bool, error) {
	query := `SELECT 1 FROM orders WHERE order_number = $1 AND user_id != $2`

	var exists int
	err := r.pool.QueryRow(ctx, query, orderNumber, userID).Scan(&exists)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *Repository) AddNewOrder(ctx context.Context, userID int, orderNumber string) (int, error) {
	query := `INSERT INTO orders (user_id, order_number) VALUES ($1, $2) RETURNING id`

	var orderID int
	err := r.pool.QueryRow(ctx, query, userID, orderNumber).Scan(&orderID)
	if err != nil {
		return 0, err
	}
	return orderID, nil
}

func (r *Repository) AddNewOrderWithCheck(ctx context.Context, userID int, orderNumber string) (int, error) {
	// Начинаем транзакцию с уровнем изоляции SERIALIZABLE для предотвращения race condition
	txOptions := pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	}
	tx, err := r.pool.BeginTx(ctx, txOptions)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// Проверяем, существует ли заказ с таким номером
	var existingUserID int
	checkQuery := `SELECT user_id FROM orders WHERE order_number = $1 FOR UPDATE`
	err = tx.QueryRow(ctx, checkQuery, orderNumber).Scan(&existingUserID)

	if err == nil {
		// Заказ уже существует
		if existingUserID == userID {
			// Заказ принадлежит тому же пользователю
			return 0, ErrOrderAlreadyExistsForUserRepo
		} else {
			// Заказ принадлежит другому пользователю
			return 0, ErrOrderAlreadyProcessedRepo
		}
	}

	if err.Error() != "no rows in result set" {
		return 0, err
	}

	// Заказа не существует, добавляем новый
	insertQuery := `INSERT INTO orders (user_id, order_number) VALUES ($1, $2) RETURNING id`
	var orderID int
	err = tx.QueryRow(ctx, insertQuery, userID, orderNumber).Scan(&orderID)
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}

	return orderID, nil
}

func (r *Repository) GetUserOrders(ctx context.Context, userID int, limit, offset int) ([]*Order, error) {
	query := `SELECT id, user_id, order_number, status, accrual, uploaded_at
			  FROM orders
			  WHERE user_id = $1
			  ORDER BY uploaded_at DESC
			  LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var order Order
		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.OrderNumber,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *Repository) GetOrdersByStatus(ctx context.Context, status string, limit int) ([]*Order, error) {
	query := `SELECT id, user_id, order_number, status, accrual, uploaded_at
			  FROM orders
			  WHERE status = $1
			  ORDER BY uploaded_at ASC
			  LIMIT $2`

	rows, err := r.pool.Query(ctx, query, status, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*Order
	for rows.Next() {
		var order Order
		if err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.OrderNumber,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *Repository) GetOrCreateOrder(ctx context.Context, userID int, orderNumber string) (int, bool, error) {
	exists, err := r.CheckUsersOrderExists(ctx, userID, orderNumber)
	if err != nil {
		return 0, false, err
	}
	if exists {
		query := `SELECT id FROM orders WHERE user_id = $1 AND order_number = $2`
		var orderID int
		err := r.pool.QueryRow(ctx, query, userID, orderNumber).Scan(&orderID)
		if err != nil {
			return 0, false, err
		}
		return orderID, true, nil
	}

	orderID, err := r.AddNewOrder(ctx, userID, orderNumber)
	if err != nil {
		return 0, false, err
	}
	return orderID, false, nil
}

func (r *Repository) UpdateOrder(ctx context.Context, userID int, orderNumber, status string, accrual float32) error {
	query := `UPDATE orders SET status = $1, accrual = $2 WHERE user_id = $3 AND order_number = $4`

	_, err := r.pool.Exec(ctx, query, status, accrual, userID, orderNumber)
	return err
}

func (r *Repository) UpdateOrderStatus(ctx context.Context, orderID int, status string) error {
	query := `UPDATE orders SET status = $1 WHERE id = $2`

	_, err := r.pool.Exec(ctx, query, status, orderID)
	return err
}
