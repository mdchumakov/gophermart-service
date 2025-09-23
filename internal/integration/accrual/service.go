package accrual

import (
	"context"
	"fmt"

	"gophermart-service/internal/config"
)

// Service представляет сервис для работы с системой ACCRUAL
type Service struct {
	client ClientInterface
	logger config.LoggerInterface
}

// NewService создает новый сервис для работы с системой ACCRUAL
func NewService(client ClientInterface, logger config.LoggerInterface) *Service {
	return &Service{
		client: client,
		logger: logger,
	}
}

// GetOrderInfo получает информацию о расчёте начислений баллов лояльности для заказа
func (s *Service) GetOrderInfo(ctx context.Context, orderNumber string) (*OrderInfo, error) {
	s.logger.Debugw("getting order info from accrual system",
		"order_number", orderNumber)

	orderInfo, err := s.client.GetOrderInfo(ctx, orderNumber)
	if err != nil {
		s.logger.Errorw("failed to get order info from accrual system",
			"order_number", orderNumber,
			"error", err)
		return nil, fmt.Errorf("failed to get order info: %w", err)
	}

	if orderInfo == nil {
		s.logger.Debugw("order not found in accrual system",
			"order_number", orderNumber)
		return nil, nil
	}

	s.logger.Debugw("successfully got order info from accrual system",
		"order_number", orderNumber,
		"status", string(orderInfo.Status),
		"accrual", orderInfo.GetAccrual())

	return orderInfo, nil
}
