package service

import (
	"gophermart-service/internal/config"
	"gophermart-service/internal/integration"
	"gophermart-service/internal/integration/accrual"
	"gophermart-service/internal/repository"
	"gophermart-service/internal/service/health"
	"gophermart-service/internal/service/jwt"
	userAuth "gophermart-service/internal/service/user/auth"
	userBalance "gophermart-service/internal/service/user/balance"
	userOrder "gophermart-service/internal/service/user/order"
	userWithdraw "gophermart-service/internal/service/user/withdraw"
)

type Services struct {
	Health       health.ServiceInterface
	UserAuth     userAuth.ServiceInterface
	UserOrder    userOrder.ServiceInterface
	UserBalance  userBalance.ServiceInterface
	UserWithdraw userWithdraw.ServiceInterface
	JWT          jwt.ServiceInterface
	Accrual      *accrual.Service
}

func NewServices(
	logger config.LoggerInterface,
	settings *config.Settings,
	repos *repository.Repositories,
	integrations *integration.Integrations,
) *Services {
	healthService := health.NewHealthService(logger, repos.Health)
	userAuthService := userAuth.NewRegisterService(logger, repos.Users)
	jwtService := jwt.NewJWTService(settings.Environment.JWT, logger)
	userOrderService := userOrder.NewOrderService(logger, repos.Orders, integrations.Accrual)
	userBalanceService := userBalance.NewUserBalanceService(logger, repos.Views)
	userWithdrawService := userWithdraw.NewUserWithdrawService(logger, repos.Orders, repos.Views, repos.Withdraw)

	return &Services{
		Health:       healthService,
		UserAuth:     userAuthService,
		UserOrder:    userOrderService,
		UserBalance:  userBalanceService,
		UserWithdraw: userWithdrawService,
		JWT:          jwtService,
	}
}
