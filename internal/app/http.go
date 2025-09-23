package app

import (
	"context"
	"gophermart-service/internal/config"
	"gophermart-service/internal/handler"
	"gophermart-service/internal/integration"
	"gophermart-service/internal/middleware"
	"gophermart-service/internal/repository"
	"gophermart-service/internal/service"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

type HTTPApp struct {
	router   *gin.Engine
	settings *config.Settings
	logger   config.LoggerInterface
	services *service.Services
	handlers *handler.Handlers
}

func NewHTTPApp(logger config.LoggerInterface) (*HTTPApp, error) {
	ctx := context.Background()

	router := gin.Default()
	settings, err := config.NewSettings()
	if err != nil {
		return nil, err
	}

	pool, err := config.SetupDB(
		ctx,
		logger,
		settings.GetDatabaseURI(),
		settings.Environment.Database.MigrationsPath,
	)
	if err != nil {
		return nil, err
	}

	repos := repository.NewRepositories(logger, pool)
	integrations := integration.NewIntegrations(logger, settings)
	services := service.NewServices(logger, settings, repos, integrations)
	handlers := handler.NewHandlers(logger, services, settings)

	return &HTTPApp{
		router:   router,
		logger:   logger,
		settings: settings,
		services: services,
		handlers: handlers,
	}, nil
}

func (a *HTTPApp) SetupCommonMiddleware() {
	a.router.Use(requestid.New())
	a.router.Use(middleware.RequestIDMiddleware())
	a.router.Use(middleware.JWTMiddleware(a.logger, a.settings.Environment.JWT, a.services.JWT))
}

func (a *HTTPApp) SetupRoutes() {
	a.router.GET("/health", a.handlers.GetHealth.Handle)
	a.router.POST("/api/user/register", a.handlers.PostUserRegister.Handle)
	a.router.POST("/api/user/login", a.handlers.PostUserLogin.Handle)
	a.router.POST("/api/user/orders", a.handlers.PostUserOrders.Handle)
	a.router.GET("/api/user/orders", a.handlers.GetUserOrders.Handle)
	a.router.GET("/api/user/balance", a.handlers.GetUserBalance.Handle)
	a.router.POST("/api/user/balance/withdraw", a.handlers.PostUserBalanceWithdraw.Handle)
	a.router.GET("/api/user/withdrawals", a.handlers.GetUserWithdrawals.Handle)
}

func (a *HTTPApp) Start() error {
	address := a.settings.GetServerAddress()

	err := a.router.Run(address)
	return err
}

func (a *HTTPApp) Stop() error {
	a.services.UserOrder.Stop()

	return nil
}
