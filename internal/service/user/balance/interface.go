package balance

import "context"

type ServiceInterface interface {
	GetBalance(ctx context.Context, userID int) (*GetUserBalanceDTO, error)
}
