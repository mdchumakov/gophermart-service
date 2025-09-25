package jwt

import "context"

type ServiceInterface interface {
	GenerateToken(ctx context.Context, user *InDTO) (string, error)
	ValidateToken(ctx context.Context, tokenString string) (*InDTO, error)
	IsTokenExpired(ctx context.Context, tokenString string) (bool, error)
	RefreshToken(ctx context.Context, tokenString string) (string, error)
}
