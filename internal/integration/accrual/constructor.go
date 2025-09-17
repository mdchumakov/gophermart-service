package accrual

import (
	"time"

	"gophermart-service/internal/config"
)

// NewClient создает новый клиент для работы с системой ACCRUAL
func NewClient(baseURL string, timeout time.Duration) ClientInterface {
	return NewHTTPClient(baseURL, timeout)
}

// NewServiceWithClient создает новый сервис с переданным клиентом
func NewServiceWithClient(client ClientInterface, logger config.LoggerInterface) *Service {
	return NewService(client, logger)
}

// NewServiceWithConfig создает новый сервис с конфигурацией
func NewServiceWithConfig(baseURL string, timeout time.Duration, logger config.LoggerInterface) *Service {
	client := NewClient(baseURL, timeout)
	return NewService(client, logger)
}
