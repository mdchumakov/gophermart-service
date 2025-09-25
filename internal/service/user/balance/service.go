package balance

import (
	"context"
	"gophermart-service/internal/base"
	"gophermart-service/internal/config"

	viewsRepo "gophermart-service/internal/repository/views"
)

func NewUserBalanceService(
	logger config.LoggerInterface,
	repo viewsRepo.RepositoryInterface,
) ServiceInterface {
	return &Service{
		logger: logger,
		repo:   repo,
	}
}

type Service struct {
	logger config.LoggerInterface
	repo   viewsRepo.RepositoryInterface
}

func (s *Service) GetBalance(ctx context.Context, userID int) (*GetUserBalanceDTO, error) {
	requestID := base.GetRequestID(ctx)

	s.logger.Infow(
		"Get user balance initiated",
		"requestID", requestID,
		"userID", userID,
	)

	balance, err := s.repo.GetUserBalance(ctx, userID)
	if err != nil {
		s.logger.Errorw(
			"Get user balance failed",
			"requestID", requestID,
			"userID", userID,
			"error", err,
		)
		return nil, err
	}

	s.logger.Infow(
		"Get user balance succeeded",
		"requestID", requestID,
		"userID", userID,
	)

	return &GetUserBalanceDTO{
		CurrentBalance: balance.CurrentBalance,
		TotalWithdrawn: balance.TotalWithdrawn,
	}, nil
}
