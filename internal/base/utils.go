package base

import (
	"context"
	"gophermart-service/internal/config"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const requestIDKey = "requestID"

func SetRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if reqID, ok := ctx.Value(requestIDKey).(string); ok {
		return reqID
	}

	return ""
}

// SetTokenToCookie устанавливает JWT токен в куки (публичная функция для использования в хендлерах)
func SetTokenToCookie(c *gin.Context, token string, jwtSettings *config.JWTSettings, expiration time.Duration) {
	// Определяем домен для куки
	domain := jwtSettings.CookieDomain
	if domain == "" {
		domain = c.Request.Host
		if strings.Contains(domain, ":") {
			domain = strings.Split(domain, ":")[0]
		}
	}

	// Устанавливаем куки
	c.SetCookie(
		jwtSettings.CookieName,
		token,
		int(expiration.Seconds()),
		jwtSettings.CookiePath,
		domain,
		jwtSettings.CookieSecure,
		jwtSettings.CookieHTTPOnly,
	)
}

func SetTokenToHeader(c *gin.Context, token string) {
	c.Header(AuthorizationHeader, BearerPrefix+token)
}
