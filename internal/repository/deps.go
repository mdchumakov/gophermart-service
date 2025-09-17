package repository

import (
	"gophermart-service/internal/config"
	"gophermart-service/internal/repository/health"
	"gophermart-service/internal/repository/orders"
	"gophermart-service/internal/repository/users"
	"gophermart-service/internal/repository/views"
	"gophermart-service/internal/repository/withdraw"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repositories struct {
	Health   health.RepositoryInterface
	Users    users.RepositoryInterface
	Orders   orders.RepositoryInterface
	Views    views.RepositoryInterface
	Withdraw withdraw.RepositoryInterface
}

func NewRepositories(logger config.LoggerInterface, pool *pgxpool.Pool) *Repositories {

	healthRepo := health.NewHealthRepository(logger, pool)
	usersRepo := users.NewUsersRepository(logger, pool)
	ordersRepo := orders.NewOrdersRepository(logger, pool)
	viewsRepo := views.NewViewsRepository(logger, pool)
	withdrawRepo := withdraw.NewWithdrawRepository(logger, pool)

	return &Repositories{
		Health:   healthRepo,
		Users:    usersRepo,
		Orders:   ordersRepo,
		Views:    viewsRepo,
		Withdraw: withdrawRepo,
	}
}
