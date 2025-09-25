package withdraw

import (
	"context"
	"errors"
	"gophermart-service/internal/base"
	"gophermart-service/internal/config"
	ordersRepo "gophermart-service/internal/repository/orders"
	viewsRepo "gophermart-service/internal/repository/views"
	withdrawRepo "gophermart-service/internal/repository/withdraw"
)

func NewUserWithdrawService(
	logger config.LoggerInterface,
	ordersRepo ordersRepo.RepositoryInterface,
	viewsRepo viewsRepo.RepositoryInterface,
	withdrawRepo withdrawRepo.RepositoryInterface,
) ServiceInterface {
	return &Service{
		logger:       logger,
		ordersRepo:   ordersRepo,
		viewsRepo:    viewsRepo,
		withdrawRepo: withdrawRepo,
	}
}

type Service struct {
	logger       config.LoggerInterface
	ordersRepo   ordersRepo.RepositoryInterface
	viewsRepo    viewsRepo.RepositoryInterface
	withdrawRepo withdrawRepo.RepositoryInterface
}

func (s *Service) MakeNewWithdraw(ctx context.Context, userID int, orderNumber string, sum float32) error {
	requestID := base.GetRequestID(ctx)

	s.logger.Infow("Load new order number initiated",
		"requestID", requestID,
		"userID", userID,
		"orderNumber", orderNumber)

	// Сначала создаем заказ (если его нет)
	_, isOrderWasCreated, err := s.ordersRepo.GetOrCreateOrder(ctx, userID, orderNumber)
	if err != nil {
		s.logger.Errorw("Make new withdraw failed - order creation",
			"requestID", requestID,
			"userID", userID,
			"orderNumber", orderNumber,
			"error", err,
		)
		return err
	}
	if isOrderWasCreated {
		s.logger.Infow("Order is already created")
	}

	// Атомарно проверяем баланс и добавляем списание в транзакции
	if err := s.withdrawRepo.AddNewWithBalanceCheck(ctx, userID, orderNumber, sum); err != nil {
		// Проверяем тип ошибки для корректной обработки
		if errors.Is(err, withdrawRepo.ErrInsufficientBalance) {
			s.logger.Warnw("Make new withdraw failed: not enough balance",
				"requestID", requestID,
				"userID", userID,
				"orderNumber", orderNumber,
				"error", "not enough balance",
			)
			return ErrNotEnoughBalance
		}

		s.logger.Errorw("Make new withdraw failed",
			"requestID", requestID,
			"userID", userID,
			"orderNumber", orderNumber,
			"error", err,
		)
		return err
	}

	return nil
}

func (s *Service) GetUserWithdrawals(ctx context.Context, userID int) ([]withdrawRepo.Withdrawal, error) {
	requestID := base.GetRequestID(ctx)

	s.logger.Infow("Get user withdrawals initiated",
		"requestID", requestID,
		"userID", userID)

	withdrawals, err := s.withdrawRepo.GetUserWithdrawals(ctx, userID)
	if err != nil {
		s.logger.Errorw("Get user withdrawals failed",
			"requestID", requestID,
			"userID", userID,
			"error", err,
		)
		return nil, err
	}

	s.logger.Infow("Get user withdrawals completed",
		"requestID", requestID,
		"userID", userID,
		"withdrawalsCount", len(withdrawals))

	return withdrawals, nil
}
