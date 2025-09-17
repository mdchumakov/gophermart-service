package health

import "context"

type RepositoryInterface interface {
	Ping(ctx context.Context) error
}
