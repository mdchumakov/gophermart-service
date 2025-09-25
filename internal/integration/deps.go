package integration

import (
	"gophermart-service/internal/config"
	"gophermart-service/internal/integration/accrual"
	"time"
)

type Integrations struct {
	Accrual accrual.ClientInterface
}

func NewIntegrations(logger config.LoggerInterface, settings *config.Settings) *Integrations {
	// Создаем ACCRUAL сервис
	accrualTimeout := time.Duration(settings.Environment.Integration.AccrualSystemTimeout) * time.Second
	accrualClient := accrual.NewServiceWithConfig(
		settings.GetAccrualSystemAddress(),
		accrualTimeout,
		logger,
	)

	return &Integrations{
		Accrual: accrualClient,
	}
}
