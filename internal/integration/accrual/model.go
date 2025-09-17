package accrual

import (
	"errors"
	"time"
)

// OrderStatus представляет статус расчёта начисления
type OrderStatus string

const (
	OrderStatusRegistered OrderStatus = "REGISTERED" // заказ зарегистрирован, но вознаграждение не рассчитано
	OrderStatusInvalid    OrderStatus = "INVALID"    // заказ не принят к расчёту, и вознаграждение не будет начислено
	OrderStatusProcessing OrderStatus = "PROCESSING" // расчёт начисления в процессе
	OrderStatusProcessed  OrderStatus = "PROCESSED"  // расчёт начисления окончен
)

// OrderInfo представляет информацию о расчёте начислений баллов лояльности
type OrderInfo struct {
	Order   string      `json:"order"`   // номер заказа
	Status  OrderStatus `json:"status"`  // статус расчёта начисления
	Accrual *float32    `json:"accrual"` // рассчитанные баллы к начислению (может отсутствовать)
}

// IsFinalStatus проверяет, является ли статус окончательным
func (s OrderStatus) IsFinalStatus() bool {
	return s == OrderStatusInvalid || s == OrderStatusProcessed
}

// HasAccrual проверяет, есть ли начисление баллов
func (o *OrderInfo) HasAccrual() bool {
	return o.Accrual != nil
}

// GetAccrual возвращает значение начисления или 0, если начисления нет
func (o *OrderInfo) GetAccrual() float32 {
	if o.Accrual == nil {
		return 0
	}
	return *o.Accrual
}

// RateLimitError представляет ошибку превышения лимита запросов
type RateLimitError struct {
	RetryAfter time.Duration
	Message    string
}

func (e *RateLimitError) Error() string {
	return e.Message
}

// IsRateLimitError проверяет, является ли ошибка связанной с превышением лимита запросов
func IsRateLimitError(err error) bool {
	var rateLimitError *RateLimitError
	ok := errors.As(err, &rateLimitError)
	return ok
}
