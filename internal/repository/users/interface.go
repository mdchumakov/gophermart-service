//go:generate mockgen -source=interface.go -destination=mock/mock_repo.go -package=mock
package users

import "context"

type RepositoryInterface interface {
	WriterRepositoryInterface
}

type WriterRepositoryInterface interface {
	Add(ctx context.Context, login, passwordHash string) (int, error)
	GetUserHashPassword(ctx context.Context, login string) (int, string, error)
}
