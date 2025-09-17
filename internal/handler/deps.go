package handler

import (
	"gophermart-service/internal/base"
	"gophermart-service/internal/config"
	"gophermart-service/internal/handler/health"
	userBalance "gophermart-service/internal/handler/user/balance"
	userLogin "gophermart-service/internal/handler/user/login"
	userOrders "gophermart-service/internal/handler/user/orders"
	userRegister "gophermart-service/internal/handler/user/register"
	userBalanceWithdraw "gophermart-service/internal/handler/user/withdraw"
	"gophermart-service/internal/service"
)

type Handlers struct {
	GetHealth               base.HandlerInterface
	PostUserRegister        base.HandlerInterface
	PostUserLogin           base.HandlerInterface
	PostUserOrders          base.HandlerInterface
	GetUserOrders           base.HandlerInterface
	GetUserBalance          base.HandlerInterface
	PostUserBalanceWithdraw base.HandlerInterface
	GetUserWithdrawals      base.HandlerInterface
}

func NewHandlers(logger config.LoggerInterface, services *service.Services, settings *config.Settings) *Handlers {
	getHealthHandler := health.NewGetHealthHandler(logger, services.Health)
	postRegisterHandler := userRegister.NewPostRegisterHandler(
		logger,
		services.UserAuth,
		services.JWT,
		settings.Environment.JWT,
	)
	postLoginHandler := userLogin.NewPostLoginHandler(
		logger,
		services.UserAuth,
		services.JWT,
		settings.Environment.JWT,
	)
	postUserOrdersHandler := userOrders.NewPostUserOrdersHandler(
		logger,
		services.UserOrder,
	)
	getUserOrdersHandler := userOrders.NewGetUserOrdersHandler(
		logger,
		services.UserOrder,
	)
	getUserBalanceHandler := userBalance.NewGetUserBalanceHandler(
		logger,
		services.UserBalance,
	)
	postUserBalanceWithdraw := userBalanceWithdraw.NewPostUserBalanceWithdraw(
		logger,
		services.UserOrder,
		services.UserWithdraw,
	)
	getUserWithdrawals := userBalanceWithdraw.NewGetUserWithdrawals(
		logger,
		services.UserWithdraw,
	)

	return &Handlers{
		GetHealth:               getHealthHandler,
		PostUserRegister:        postRegisterHandler,
		PostUserLogin:           postLoginHandler,
		PostUserOrders:          postUserOrdersHandler,
		GetUserOrders:           getUserOrdersHandler,
		GetUserBalance:          getUserBalanceHandler,
		PostUserBalanceWithdraw: postUserBalanceWithdraw,
		GetUserWithdrawals:      getUserWithdrawals,
	}
}
