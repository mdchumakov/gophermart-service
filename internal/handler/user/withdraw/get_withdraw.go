package withdraw

import (
	"gophermart-service/internal/base"
	"gophermart-service/internal/config"
	"gophermart-service/internal/service/jwt"
	serviceUserWithdraw "gophermart-service/internal/service/user/withdraw"
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

type getUserWithdrawals struct {
	logger              config.LoggerInterface
	userWithdrawService serviceUserWithdraw.ServiceInterface
}

func NewGetUserWithdrawals(
	logger config.LoggerInterface,
	userWithdrawService serviceUserWithdraw.ServiceInterface,
) base.HandlerInterface {
	return &getUserWithdrawals{
		logger:              logger,
		userWithdrawService: userWithdrawService,
	}
}

func (h *getUserWithdrawals) Handle(c *gin.Context) {
	requestID := requestid.Get(c)

	user := jwt.ExtractUserFromContext(c.Request.Context())
	if user == nil {
		h.logger.Warnw("user not found in context", "requestID", requestID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	withdrawals, err := h.userWithdrawService.GetUserWithdrawals(c.Request.Context(), user.ID)
	if err != nil {
		h.logger.Errorw("failed to get user withdrawals", "requestID", requestID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if len(withdrawals) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, withdrawals)
}
