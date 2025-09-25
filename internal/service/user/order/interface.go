package order

import "context"

type ServiceInterface interface {
	LoadNewOrderNumber(ctx context.Context, userID int, orderNumber string) error
	ValidateOrderNumber(ctx context.Context, orderNumber string) error
	GetUserOrders(ctx context.Context, userID int, limit, offset int) ([]*orderDTO, error)
	Stop()
}
