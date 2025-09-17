package withdraw

import (
	"context"
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

func (s *Service) MakeNewWithdraw(ctx context.Context, userId int, orderNumber string, sum float32) error {
	requestID := base.GetRequestID(ctx)

	s.logger.Infow("Load new order number initiated",
		"requestID", requestID,
		"userID", userId,
		"orderNumber", orderNumber)

	_, isOrderWasCreated, err := s.ordersRepo.GetOrCreateOrder(ctx, userId, orderNumber)
	if err != nil {
		s.logger.Errorw("Make new withdraw failed",
			"requestID", requestID,
			"userID", userId,
			"orderNumber", orderNumber,
			"error", err,
		)
		return err
	}
	if isOrderWasCreated {
		s.logger.Infow("Order is already created")
	}

	currentBalance, err := s.viewsRepo.GetUserBalance(ctx, userId)
	if err != nil {
		return err
	}

	if currentBalance.CurrentBalance < sum {
		s.logger.Warnw("Make new withdraw failed: not enough balance",
			"requestID", requestID,
			"userID", userId,
			"orderNumber", orderNumber,
			"error", "not enough balance",
		)
		return ErrNotEnoughBalance
	}

	if err := s.withdrawRepo.AddNew(ctx, userId, orderNumber, sum); err != nil {
		s.logger.Errorw("Make new withdraw failed",
			"requestID", requestID,
			"userID", userId,
			"orderNumber", orderNumber,
			"error", err,
		)
		return err
	}

	return nil
}

func (s *Service) GetUserWithdrawals(ctx context.Context, userId int) ([]withdrawRepo.Withdrawal, error) {
	requestID := base.GetRequestID(ctx)

	s.logger.Infow("Get user withdrawals initiated",
		"requestID", requestID,
		"userID", userId)

	withdrawals, err := s.withdrawRepo.GetUserWithdrawals(ctx, userId)
	if err != nil {
		s.logger.Errorw("Get user withdrawals failed",
			"requestID", requestID,
			"userID", userId,
			"error", err,
		)
		return nil, err
	}

	s.logger.Infow("Get user withdrawals completed",
		"requestID", requestID,
		"userID", userId,
		"withdrawalsCount", len(withdrawals))

	return withdrawals, nil
}
