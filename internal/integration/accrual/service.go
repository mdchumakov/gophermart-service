package accrual

import (
	"context"
	"fmt"
	"time"

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

// WaitForOrderProcessing ожидает завершения обработки заказа в системе ACCRUAL
// с периодическими проверками статуса
func (s *Service) WaitForOrderProcessing(ctx context.Context, orderNumber string,
	checkInterval time.Duration, maxWaitTime time.Duration) (*OrderInfo, error) {

	s.logger.Infow("waiting for order processing in accrual system",
		"order_number", orderNumber,
		"check_interval", checkInterval,
		"max_wait_time", maxWaitTime)

	timeoutCtx, cancel := context.WithTimeout(ctx, maxWaitTime)
	defer cancel()

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			s.logger.Warnw("timeout waiting for order processing",
				"order_number", orderNumber)
			return nil, fmt.Errorf("timeout waiting for order processing: %w", timeoutCtx.Err())

		case <-ticker.C:
			orderInfo, err := s.GetOrderInfo(ctx, orderNumber)
			if err != nil {
				// Если это ошибка превышения лимита запросов, продолжаем ожидание
				if IsRateLimitError(err) {
					s.logger.Warnw("rate limit exceeded, continuing to wait",
						"order_number", orderNumber)
					continue
				}
				return nil, err
			}

			if orderInfo == nil {
				s.logger.Debugw("order not found, continuing to wait",
					"order_number", orderNumber)
				continue
			}

			// Если статус окончательный, возвращаем результат
			if orderInfo.Status.IsFinalStatus() {
				s.logger.Infow("order processing completed",
					"order_number", orderNumber,
					"status", string(orderInfo.Status))
				return orderInfo, nil
			}

			s.logger.Debugw("order still processing, continuing to wait",
				"order_number", orderNumber,
				"status", string(orderInfo.Status))
		}
	}
}
