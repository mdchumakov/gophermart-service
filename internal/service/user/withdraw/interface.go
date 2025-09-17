package withdraw

import (
	"context"
	withdrawRepo "gophermart-service/internal/repository/withdraw"
)

type ServiceInterface interface {
	MakeNewWithdraw(ctx context.Context, userId int, orderNumber string, sum float32) error
	GetUserWithdrawals(ctx context.Context, userId int) ([]withdrawRepo.Withdrawal, error)
}
