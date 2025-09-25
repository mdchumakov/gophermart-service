package withdraw

import (
	"gophermart-service/internal/base"
	"gophermart-service/internal/config"
	"gophermart-service/internal/service/jwt"
	serviceUserOrders "gophermart-service/internal/service/user/order"
	serviceUserWithdraw "gophermart-service/internal/service/user/withdraw"
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

type postUserBalanceWithdraw struct {
	logger              config.LoggerInterface
	userOrdersService   serviceUserOrders.ServiceInterface
	userWithdrawService serviceUserWithdraw.ServiceInterface
}

type RequestBody struct {
	OrderNumber string  `json:"order"`
	Sum         float32 `json:"sum"`
}

func NewPostUserBalanceWithdraw(
	logger config.LoggerInterface,
	userOrdersService serviceUserOrders.ServiceInterface,
	userWithdrawService serviceUserWithdraw.ServiceInterface,
) base.HandlerInterface {
	return &postUserBalanceWithdraw{
		logger:              logger,
		userOrdersService:   userOrdersService,
		userWithdrawService: userWithdrawService,
	}
}

func (h *postUserBalanceWithdraw) Handle(c *gin.Context) {
	requestID := requestid.Get(c)

	user := jwt.ExtractUserFromContext(c.Request.Context())
	if user == nil {
		h.logger.Warnw("user not found in context", "requestID", requestID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var requestBody RequestBody
	if err := c.ShouldBindBodyWithJSON(&requestBody); err != nil {
		h.logger.Warnw("failed to bind request body", "requestID", requestID, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.userOrdersService.ValidateOrderNumber(c.Request.Context(), requestBody.OrderNumber); err != nil {
		h.logger.Warnw("invalid order number", "requestID", requestID, "error", err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid order number"})
		return
	}

	if err := h.userWithdrawService.MakeNewWithdraw(
		c.Request.Context(),
		user.ID,
		requestBody.OrderNumber,
		requestBody.Sum); err != nil {
		if serviceUserWithdraw.IsErrNotEnoughBalance(err) {
			h.logger.Warnw("not enough balance for withdraw", "requestID", requestID, "error", err)
			c.JSON(http.StatusPaymentRequired, gin.H{"error": "not enough balance"})
			return
		}
		h.logger.Errorw("failed to make new withdraw", "requestID", requestID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Status(http.StatusOK)
}
