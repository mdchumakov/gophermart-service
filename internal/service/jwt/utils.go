package jwt

import "context"

const TokenContextKey = "jwtToken"

func SetUserInContext(ctx context.Context, user *InDTO) context.Context {
	return context.WithValue(ctx, TokenContextKey, user)
}

func ExtractUserFromContext(ctx context.Context) *InDTO {
	if ctx == nil {
		return nil
	}

	if user, ok := ctx.Value(TokenContextKey).(*InDTO); ok {
		return user
	}

	return nil
}
