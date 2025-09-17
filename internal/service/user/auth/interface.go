package auth

import "context"

type ServiceInterface interface {
	RegisterUser(ctx context.Context, dtoIn *InDTO) (*OutDTO, error)
	LoginUser(ctx context.Context, dtoIn *InDTO) (*OutDTO, error)
	HashPassword(password string) (string, error)
}
