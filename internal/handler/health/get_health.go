package health

import (
	"gophermart-service/internal/base"
	"gophermart-service/internal/config"
	serviceHealth "gophermart-service/internal/service/health"
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

type getHealthHandler struct {
	logger  config.LoggerInterface
	service serviceHealth.ServiceInterface
}

func NewGetHealthHandler(
	logger config.LoggerInterface,
	service serviceHealth.ServiceInterface,
) base.HandlerInterface {
	return &getHealthHandler{
		logger:  logger,
		service: service,
	}
}

func (h *getHealthHandler) Handle(c *gin.Context) {
	requestID := requestid.Get(c)

	h.logger.Infow("Starting health check", "requestID", requestID)
	if err := h.service.Check(c.Request.Context()); err != nil {
		h.logger.Errorw("Health check failed",
			"error", err,
		)
		c.String(http.StatusInternalServerError, `{"error": "health check failed"}`)
		return
	}

	c.String(http.StatusOK, `OK`)

	h.logger.Infow("Health check done", "requestID", requestID)
}
