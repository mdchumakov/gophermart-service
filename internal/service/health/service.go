package health

import (
	"context"
	"gophermart-service/internal/base"
	"gophermart-service/internal/config"
	healthRepo "gophermart-service/internal/repository/health"
)

func NewHealthService(
	logger config.LoggerInterface,
	repo healthRepo.RepositoryInterface,
) ServiceInterface {
	return &Service{
		logger: logger,
		repo:   repo,
	}
}

type Service struct {
	logger config.LoggerInterface
	repo   healthRepo.RepositoryInterface
}

func (s *Service) Check(ctx context.Context) error {
	requestID := base.GetRequestID(ctx)

	s.logger.Infow("Health check initiated", "requestID", requestID)

	err := s.repo.Ping(ctx)
	if err != nil {
		return err
	}

	s.logger.Infow("Health check completed", "requestID", requestID)
	return nil
}
