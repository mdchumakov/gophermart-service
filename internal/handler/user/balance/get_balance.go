package balance

import (
	"gophermart-service/internal/base"
	"gophermart-service/internal/config"
	"gophermart-service/internal/service/jwt"
	userBalance "gophermart-service/internal/service/user/balance"
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

type getUserBalanceHandler struct {
	logger             config.LoggerInterface
	userBalanceService userBalance.ServiceInterface
}

func NewGetUserBalanceHandler(
	logger config.LoggerInterface,
	userBalanceService userBalance.ServiceInterface,
) base.HandlerInterface {
	return &getUserBalanceHandler{
		logger:             logger,
		userBalanceService: userBalanceService,
	}
}

func (h *getUserBalanceHandler) Handle(c *gin.Context) {
	requestID := requestid.Get(c)

	user := jwt.ExtractUserFromContext(c.Request.Context())
	if user == nil {
		h.logger.Warnw("user not found in context", "requestID", requestID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	h.logger.Infow("Starting to handle get user balance", "requestID", requestID)

	balance, err := h.userBalanceService.GetBalance(c.Request.Context(), user.ID)
	if err != nil {
		h.logger.Errorw("Failed to get user balance", "requestID", requestID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, balance)
}
