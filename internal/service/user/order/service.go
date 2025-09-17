package order

import (
	"context"
	"gophermart-service/internal/base"
	"gophermart-service/internal/config"
	"gophermart-service/internal/integration/accrual"
	ordersRepo "gophermart-service/internal/repository/orders"
	"strconv"
	"strings"
	"sync"
)

const (
	maxWorkers         = 3
	maxChannelPoolSize = 100
)

func NewOrderService(
	logger config.LoggerInterface,
	repo ordersRepo.RepositoryInterface,
	accrualClient accrual.ClientInterface,
) ServiceInterface {

	service := &Service{
		logger:        logger,
		repo:          repo,
		waitGroup:     &sync.WaitGroup{},
		accrualClient: accrualClient,
		addChan:       make(chan addNewOrderRequest, maxChannelPoolSize),
		stopChan:      make(chan struct{}),
	}

	for workerID := 1; workerID <= maxWorkers; workerID++ {
		service.waitGroup.Add(1)
		go service.addNewOrderWorker(workerID)
	}

	return service
}

type Service struct {
	logger        config.LoggerInterface
	repo          ordersRepo.RepositoryInterface
	accrualClient accrual.ClientInterface
	waitGroup     *sync.WaitGroup
	addChan       chan addNewOrderRequest
	stopChan      chan struct{}
}

type addNewOrderRequest struct {
	UserID      int
	OrderNumber string
	requestID   string
	logger      config.LoggerInterface
}

func (s *Service) LoadNewOrderNumber(ctx context.Context, userId int, orderNumber string) error {
	requestID := base.GetRequestID(ctx)

	s.logger.Infow("Load new order number initiated",
		"requestID", requestID,
		"userID", userId,
		"orderNumber", orderNumber)

	if err := s.ValidateOrderNumber(ctx, orderNumber); err != nil {
		if IsErrBadOrderNumber(err) {
			return ErrBadOrderNumber
		}
		return err
	}

	isUsersOrderAlreadyExists, err := s.repo.CheckUsersOrderExists(ctx, userId, orderNumber)
	if err != nil {
		return err
	}
	if isUsersOrderAlreadyExists {
		s.logger.Warnw("Order number already exists for this user",
			"requestID", requestID,
			"userID", userId,
			"orderNumber", orderNumber)
		return ErrOrderAlreadyExistsForUser
	}

	isOrderAlreadyProcessed, err := s.repo.CheckOrderAlreadyProcessed(ctx, userId, orderNumber)
	if err != nil {
		return err
	}
	if isOrderAlreadyProcessed {
		s.logger.Warnw("Order number already processed by another user",
			"requestID", requestID,
			"userID", userId,
			"orderNumber", orderNumber)
		return ErrOrderAlreadyProcessed
	}

	select {
	case s.addChan <- addNewOrderRequest{
		logger:      s.logger,
		requestID:   requestID,
		UserID:      userId,
		OrderNumber: orderNumber,
	}:
		s.logger.Debugw("Order addition request sent to worker",
			"requestID", requestID,
			"userID", userId,
			"orderNumber", orderNumber)
	default:
		s.logger.Warnw("Add order channel is full, processing synchronously",
			"requestID", requestID,
			"userID", userId,
			"orderNumber", orderNumber,
		)
		return ErrTooManyOrders
	}

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

func (s *Service) GetUserOrders(ctx context.Context, userId int, limit, offset int) ([]*orderDTO, error) {
	requestID := base.GetRequestID(ctx)

	s.logger.Infow("Get user orders request",
		"requestID", requestID,
		"userID", userId,
	)

	orders, err := s.repo.GetUserOrders(ctx, userId, limit, offset)
	if err != nil {
		return nil, err
	}

	if len(orders) == 0 {
		s.logger.Infow("No orders found for user",
			"requestID", requestID,
			"userID", userId,
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
		"userID", userId,
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

func (s *Service) addNewOrderWorker(workerID int) {
	defer s.waitGroup.Done()

	for {
		select {
		case req := <-s.addChan:
			ctx := context.Background()

			req.logger.Debugw("Worker started processing delete request",
				"worker_id", workerID,
				"request_id", req.requestID,
				"order_number", req.OrderNumber,
				"user_id", req.UserID,
			)

			orderID, err := s.repo.AddNewOrder(ctx, req.UserID, req.OrderNumber)
			if err != nil {
				s.logger.Errorw("Failed to add new order",
					"requestID", req.requestID,
					"userID", req.UserID,
					"orderNumber", req.OrderNumber,
					"error", err.Error())
			}
			if orderID == 0 {
				s.logger.Errorw("Failed to add new order",
					"requestID", req.requestID,
					"userID", req.UserID,
					"orderNumber", req.OrderNumber,
				)
			}

			orderInfo, err := s.accrualClient.GetOrderInfo(ctx, req.OrderNumber)
			if err != nil {
				s.logger.Errorw("Failed to get order info",
					"err", err.Error(),
					"worker_id", workerID,
					"request_id", req.requestID,
					"order_number", req.OrderNumber,
					"user_id", req.UserID,
				)
				return
			}
			if orderInfo == nil {
				s.logger.Warnw("Failed to get order info",
					"worker_id", workerID,
					"request_id", req.requestID,
					"order_number", req.OrderNumber,
					"user_id", req.UserID,
				)
				return
			}

			if err = s.repo.UpdateOrder(
				ctx,
				req.UserID,
				req.OrderNumber,
				string(orderInfo.Status),
				orderInfo.GetAccrual(),
			); err != nil {
				s.logger.Errorw("Failed to update order",
					"err", err.Error(),
					"worker_id", workerID,
					"request_id", req.requestID,
					"order_number", req.OrderNumber,
					"user_id", req.UserID,
				)
				return
			}

			req.logger.Debugw("Worker finished processing delete request",
				"worker_id", workerID,
				"request_id", req.requestID,
				"order_number", req.OrderNumber,
				"user_id", req.UserID,
			)

		case <-s.stopChan:
			return
		}
	}
}

func (s *Service) Stop() {}
