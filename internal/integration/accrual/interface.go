package accrual

import "context"

// ClientInterface представляет интерфейс для работы с системой ACCRUAL
type ClientInterface interface {
	// GetOrderInfo получает информацию о расчёте начислений баллов лояльности для заказа
	GetOrderInfo(ctx context.Context, orderNumber string) (*OrderInfo, error)
}
