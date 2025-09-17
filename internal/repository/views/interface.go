package views

import "context"

type RepositoryInterface interface {
	RepositoryReaderInterface
}

type RepositoryReaderInterface interface {
	GetUserBalance(ctx context.Context, userID int) (*UserBalance, error)
}
