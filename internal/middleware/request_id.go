package middleware

import (
	"gophermart-service/internal/base"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// RequestIDMiddleware Middleware для добавления requestID в context
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := requestid.Get(c)
		ctx := base.SetRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
