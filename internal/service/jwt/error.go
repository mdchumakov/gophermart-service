package jwt

import (
	"errors"
	"gophermart-service/internal/base"
)

var (
	ErrUserCannotBeNil          = errors.New("user cannot be nil")
	ErrTokenCannotBeEmpty       = errors.New("token cannot be empty")
	ErrInvalidTokenFormat       = errors.New("invalid token format")
	ErrInvalidTokenClaims       = errors.New("invalid token claims")
	ErrTokenHasNoExpirationTime = errors.New("token has no expiration time")
)

type ErrTokenSigning struct {
	base.Exception
}

type ErrFailedToParseToken struct {
	base.Exception
}

type ErrUnexpectedSigningMethod struct {
	base.Exception
}
