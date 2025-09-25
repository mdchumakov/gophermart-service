package health

import "context"

type ServiceInterface interface {
	Check(ctx context.Context) error
}
