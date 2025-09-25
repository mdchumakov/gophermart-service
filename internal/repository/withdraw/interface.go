package withdraw

import "context"

type RepositoryInterface interface {
	RepositoryWriterInterface
	RepositoryReaderInterface
}

type RepositoryWriterInterface interface {
	AddNew(ctx context.Context, userID int, orderNumber string, sum float32) error
	AddNewWithBalanceCheck(ctx context.Context, userID int, orderNumber string, sum float32) error
}

type RepositoryReaderInterface interface {
	GetUserWithdrawals(ctx context.Context, userID int) ([]Withdrawal, error)
}
