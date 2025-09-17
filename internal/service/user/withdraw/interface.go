package withdraw

import (
	"context"
	withdrawRepo "gophermart-service/internal/repository/withdraw"
)

type ServiceInterface interface {
	MakeNewWithdraw(ctx context.Context, userID int, orderNumber string, sum float32) error
	GetUserWithdrawals(ctx context.Context, userID int) ([]withdrawRepo.Withdrawal, error)
}
