package middleware

import (
	"gophermart-service/internal/base"
	"gophermart-service/internal/config"
	"gophermart-service/internal/service/jwt"
	"net/http"
	"strings"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

func JWTMiddleware(
	logger config.LoggerInterface,
	jwtSettings *config.JWTSettings,
	jwtService jwt.ServiceInterface,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := requestid.Get(c)
		requestCtx := c.Request.Context()

		token := ExtractToken(c, logger, jwtSettings.CookieName)
		if token == "" {
			logger.Warnw("JWT token is missing", "request_id", requestID)
			return
		}

		user, err := jwtService.ValidateToken(requestCtx, token)
		if err != nil {
			logger.Warnw("Invalid JWT token", "error", err, "request_id", requestID)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// Добавляем информацию о пользователе в контекст запроса
		requestCtx = jwt.SetUserInContext(requestCtx, user)
		c.Request = c.Request.WithContext(requestCtx)
	}
}

func ExtractToken(c *gin.Context, logger config.LoggerInterface, cookieName string) string {
	requestID := requestid.Get(c)

	token := ExtractTokenFromCookie(c, cookieName)
	if token != "" {
		logger.Debugw("JWT token found in request Cookie", "request_id", requestID)
		return token
	}
	token = ExtractTokenFromHeader(c)
	if token != "" {
		logger.Debugw("JWT token found in request Header", "request_id", requestID)
		return token
	}
	logger.Debugw("No JWT token found in request", "request_id", requestID)
	return ""
}

func ExtractTokenFromCookie(c *gin.Context, cookieName string) string {
	token, err := c.Cookie(cookieName)
	if err != nil {
		return ""
	}

	if token == "" {
		return ""
	}
	return token
}

// ExtractTokenFromHeader извлекает JWT токен из заголовка Authorization
func ExtractTokenFromHeader(c *gin.Context) string {
	authHeader := c.GetHeader(base.AuthorizationHeader)
	if authHeader == "" {
		return ""
	}

	// Проверяем формат "Bearer <token>"
	if !strings.HasPrefix(authHeader, base.BearerPrefix) {
		return ""
	}

	// Извлекаем токен (убираем префикс "Bearer ")
	token := strings.TrimPrefix(authHeader, base.BearerPrefix)
	if token == "" {
		return ""
	}

	return token
}
