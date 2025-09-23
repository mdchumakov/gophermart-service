package orders

import "context"

type RepositoryInterface interface {
	ReaderRepositoryInterface
	WriterRepositoryInterface

	GetOrCreateOrder(ctx context.Context, userID int, orderNumber string) (int, bool, error)
}

type ReaderRepositoryInterface interface {
	CheckUsersOrderExists(ctx context.Context, userID int, orderNumber string) (bool, error)
	CheckOrderAlreadyProcessed(ctx context.Context, userID int, orderNumber string) (bool, error)
	GetUserOrders(ctx context.Context, userID int, limit, offset int) ([]*Order, error)
	GetOrdersByStatus(ctx context.Context, status string, limit int) ([]*Order, error)
}

type WriterRepositoryInterface interface {
	AddNewOrder(ctx context.Context, userID int, orderNumber string) (int, error)
	AddNewOrderWithCheck(ctx context.Context, userID int, orderNumber string) (int, error)
	UpdateOrder(ctx context.Context, userID int, orderNumber, status string, accrual float32) error
	UpdateOrderStatus(ctx context.Context, orderID int, status string) error
}
