package order

import (
	"context"
	"errors"
	"gophermart-service/internal/base"
	"gophermart-service/internal/config"
	"gophermart-service/internal/integration/accrual"
	ordersRepo "gophermart-service/internal/repository/orders"
	"strconv"
	"strings"
	"time"
)

const (
	maxWorkers         = 3
	batchSize          = 10
	processingInterval = 5 * time.Second
)

func NewOrderService(
	logger config.LoggerInterface,
	repo ordersRepo.RepositoryInterface,
	accrualClient accrual.ClientInterface,
) ServiceInterface {

	service := &Service{
		logger:        logger,
		repo:          repo,
		accrualClient: accrualClient,
		stopChan:      make(chan struct{}),
		rateLimitChan: make(chan time.Duration, 1), // буферизованный канал для rate limiting
	}

	// Запускаем воркеры для батчевой обработки заказов
	for workerID := 1; workerID <= maxWorkers; workerID++ {
		go service.batchOrderProcessor(workerID)
	}

	return service
}

type Service struct {
	logger        config.LoggerInterface
	repo          ordersRepo.RepositoryInterface
	accrualClient accrual.ClientInterface
	stopChan      chan struct{}
	rateLimitChan chan time.Duration // канал для передачи времени ожидания при rate limiting
}

func (s *Service) LoadNewOrderNumber(ctx context.Context, userID int, orderNumber string) error {
	requestID := base.GetRequestID(ctx)

	s.logger.Infow("Load new order number initiated",
		"requestID", requestID,
		"userID", userID,
		"orderNumber", orderNumber)

	if err := s.ValidateOrderNumber(ctx, orderNumber); err != nil {
		if IsErrBadOrderNumber(err) {
			return ErrBadOrderNumber
		}
		return err
	}

	// Атомарно добавляем заказ в БД с проверкой уникальности в транзакции
	orderID, err := s.repo.AddNewOrderWithCheck(ctx, userID, orderNumber)
	if err != nil {
		if errors.Is(err, ordersRepo.ErrOrderAlreadyExistsForUserRepo) {
			s.logger.Warnw("Order number already exists for this user",
				"requestID", requestID,
				"userID", userID,
				"orderNumber", orderNumber)
			return ErrOrderAlreadyExistsForUser
		}
		if errors.Is(err, ordersRepo.ErrOrderAlreadyProcessedRepo) {
			s.logger.Warnw("Order number already processed by another user",
				"requestID", requestID,
				"userID", userID,
				"orderNumber", orderNumber)
			return ErrOrderAlreadyProcessed
		}

		s.logger.Errorw("Failed to add new order",
			"requestID", requestID,
			"userID", userID,
			"orderNumber", orderNumber,
			"error", err.Error())
		return ErrFailedToAddOrder
	}
	if orderID == 0 {
		s.logger.Errorw("Failed to add new order - zero order ID",
			"requestID", requestID,
			"userID", userID,
			"orderNumber", orderNumber)
		return ErrFailedToAddOrder
	}

	s.logger.Infow("Order added to database successfully",
		"requestID", requestID,
		"userID", userID,
		"orderNumber", orderNumber,
		"orderID", orderID)

	return nil
}

func (s *Service) ValidateOrderNumber(ctx context.Context, orderNumber string) error {
	requestID := base.GetRequestID(ctx)

	if s.luhnCheck(orderNumber) {
		s.logger.Infow("Order number is valid", "requestID", requestID, "orderNumber", orderNumber)
		return nil
	}

	s.logger.Infow("Order number is invalid", "requestID", requestID, "orderNumber", orderNumber)
	return ErrBadOrderNumber
}

func (s *Service) GetUserOrders(ctx context.Context, userID int, limit, offset int) ([]*orderDTO, error) {
	requestID := base.GetRequestID(ctx)

	s.logger.Infow("Get user orders request",
		"requestID", requestID,
		"userID", userID,
	)

	orders, err := s.repo.GetUserOrders(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	if len(orders) == 0 {
		s.logger.Infow("No orders found for user",
			"requestID", requestID,
			"userID", userID,
		)
		return nil, ErrNoOrders
	}

	var result []*orderDTO
	for _, order := range orders {
		result = append(result, &orderDTO{
			OrderNumber: order.OrderNumber,
			Status:      order.Status,
			Accrual:     order.Accrual,
			UploadedAt:  order.UploadedAt,
		})
	}

	s.logger.Infow("User orders retrieved successfully",
		"requestID", requestID,
		"userID", userID,
		"orders_count", len(result),
	)

	return result, nil
}

func (s *Service) luhnCheck(number string) bool {
	// Удаляем пробелы и дефисы
	number = strings.ReplaceAll(number, " ", "")
	number = strings.ReplaceAll(number, "-", "")

	// Проверяем, что строка содержит только цифры
	if len(number) == 0 {
		return false
	}

	sum := 0
	isEven := false

	// Проходим по цифрам справа налево
	for i := len(number) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false // Не цифра
		}

		if isEven {
			digit *= 2
			if digit > 9 {
				digit = digit/10 + digit%10
			}
		}

		sum += digit
		isEven = !isEven
	}

	return sum%10 == 0
}

func (s *Service) batchOrderProcessor(workerID int) {
	s.logger.Infow("Batch order processor started", "worker_id", workerID)
	defer s.logger.Infow("Batch order processor stopped", "worker_id", workerID)

	ticker := time.NewTicker(processingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.processBatchOrders(workerID)

		case retryAfter := <-s.rateLimitChan:
			s.logger.Warnw("Rate limit detected, sleeping all workers",
				"worker_id", workerID,
				"retry_after", retryAfter)
			s.sleepWithGracefulShutdown(workerID, retryAfter)

		case <-s.stopChan:
			s.logger.Infow("Batch order processor received stop signal", "worker_id", workerID)
			return
		}
	}
}

func (s *Service) processBatchOrders(workerID int) {
	ctx := context.Background()

	// Получаем заказы со статусом NEW
	orders, err := s.repo.GetOrdersByStatus(ctx, "NEW", batchSize)
	if err != nil {
		s.logger.Errorw("Failed to get orders by status",
			"worker_id", workerID,
			"status", "NEW",
			"error", err.Error())
		return
	}

	if len(orders) == 0 {
		s.logger.Debugw("No NEW orders to process", "worker_id", workerID)
		return
	}

	s.logger.Infow("Processing batch of orders",
		"worker_id", workerID,
		"orders_count", len(orders))

	for _, order := range orders {
		s.processOrder(workerID, order)
	}
}

func (s *Service) sleepWithGracefulShutdown(workerID int, duration time.Duration) {
	s.logger.Infow("Worker sleeping due to rate limit",
		"worker_id", workerID,
		"duration", duration)

	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-timer.C:
		s.logger.Infow("Worker woke up from rate limit sleep",
			"worker_id", workerID)
		// Продолжаем работу

	case <-s.stopChan:
		s.logger.Infow("Worker received stop signal during rate limit sleep",
			"worker_id", workerID)
		return
	}
}

func (s *Service) processOrder(workerID int, order *ordersRepo.Order) {
	ctx := context.Background()

	s.logger.Debugw("Processing order",
		"worker_id", workerID,
		"order_id", order.ID,
		"order_number", order.OrderNumber,
		"user_id", order.UserID)

	// Обновляем статус на PROCESSING
	if err := s.repo.UpdateOrderStatus(ctx, order.ID, "PROCESSING"); err != nil {
		s.logger.Errorw("Failed to update order status to PROCESSING",
			"worker_id", workerID,
			"order_id", order.ID,
			"order_number", order.OrderNumber,
			"error", err.Error())
		return
	}

	// Получаем информацию о заказе из системы accrual
	orderInfo, err := s.accrualClient.GetOrderInfo(ctx, order.OrderNumber)
	if err != nil {
		// Проверяем, является ли это ошибкой rate limiting
		if accrual.IsRateLimitError(err) {
			var rateLimitErr *accrual.RateLimitError
			if errors.As(err, &rateLimitErr) {
				s.logger.Warnw("Rate limit exceeded, signaling all workers to sleep",
					"worker_id", workerID,
					"order_id", order.ID,
					"order_number", order.OrderNumber,
					"retry_after", rateLimitErr.RetryAfter)

				// Отправляем сигнал rate limiting всем воркерам (неблокирующе)
				select {
				case s.rateLimitChan <- rateLimitErr.RetryAfter:
				default:
				}

				// Возвращаем статус обратно на NEW для повторной обработки
				if updateErr := s.repo.UpdateOrderStatus(ctx, order.ID, "NEW"); updateErr != nil {
					s.logger.Errorw("Failed to revert order status to NEW",
						"worker_id", workerID,
						"order_id", order.ID,
						"error", updateErr.Error())
				}
				return
			}
		}

		s.logger.Errorw("Failed to get order info from accrual system",
			"worker_id", workerID,
			"order_id", order.ID,
			"order_number", order.OrderNumber,
			"error", err.Error())
		// Возвращаем статус обратно на NEW для повторной обработки
		if updateErr := s.repo.UpdateOrderStatus(ctx, order.ID, "NEW"); updateErr != nil {
			s.logger.Errorw("Failed to revert order status to NEW",
				"worker_id", workerID,
				"order_id", order.ID,
				"error", updateErr.Error())
		}
		return
	}

	if orderInfo == nil {
		s.logger.Warnw("Order not found in accrual system",
			"worker_id", workerID,
			"order_id", order.ID,
			"order_number", order.OrderNumber)
		// Возвращаем статус обратно на NEW для повторной обработки
		if updateErr := s.repo.UpdateOrderStatus(ctx, order.ID, "NEW"); updateErr != nil {
			s.logger.Errorw("Failed to revert order status to NEW",
				"worker_id", workerID,
				"order_id", order.ID,
				"error", updateErr.Error())
		}
		return
	}

	// Обновляем заказ с финальным статусом и начислением
	if err = s.repo.UpdateOrder(
		ctx,
		order.UserID,
		order.OrderNumber,
		string(orderInfo.Status),
		orderInfo.GetAccrual(),
	); err != nil {
		s.logger.Errorw("Failed to update order with final status",
			"worker_id", workerID,
			"order_id", order.ID,
			"order_number", order.OrderNumber,
			"error", err.Error())
		// Возвращаем статус обратно на NEW для повторной обработки
		if updateErr := s.repo.UpdateOrderStatus(ctx, order.ID, "NEW"); updateErr != nil {
			s.logger.Errorw("Failed to revert order status to NEW",
				"worker_id", workerID,
				"order_id", order.ID,
				"error", updateErr.Error())
		}
		return
	}

	s.logger.Infow("Order processed successfully",
		"worker_id", workerID,
		"order_id", order.ID,
		"order_number", order.OrderNumber,
		"user_id", order.UserID,
		"status", string(orderInfo.Status),
		"accrual", orderInfo.GetAccrual())
}

func (s *Service) Stop() {
	s.logger.Info("Stopping order service...")

	// Отправляем сигнал остановки всем воркерам
	close(s.stopChan)

	// Закрываем канал rate limiting
	close(s.rateLimitChan)

	s.logger.Info("Order service stopped")
}
