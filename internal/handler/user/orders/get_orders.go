package orders

import (
	"gophermart-service/internal/base"
	"gophermart-service/internal/config"
	"gophermart-service/internal/service/jwt"
	serviceUserOrders "gophermart-service/internal/service/user/order"
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

type getUserOrdersHandler struct {
	logger            config.LoggerInterface
	userOrdersService serviceUserOrders.ServiceInterface
}

type PaginationParams struct {
	Limit  int `form:"limit" binding:"min=1,max=100"`
	Offset int `form:"offset" binding:"min=0"`
}

func NewGetUserOrdersHandler(
	logger config.LoggerInterface,
	userOrdersService serviceUserOrders.ServiceInterface,
) base.HandlerInterface {
	return &getUserOrdersHandler{
		logger:            logger,
		userOrdersService: userOrdersService,
	}
}

func (h *getUserOrdersHandler) Handle(c *gin.Context) {
	requestID := requestid.Get(c)

	user := jwt.ExtractUserFromContext(c.Request.Context())
	if user == nil {
		h.logger.Warnw("user not found in context", "requestID", requestID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var params PaginationParams
	params.Limit = 100
	params.Offset = 0

	h.logger.Infow("Starting to handle get user orders", "requestID", requestID)

	if err := c.ShouldBindQuery(&params); err != nil {
		h.logger.Warnw("Invalid pagination parameters", "requestID", requestID, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pagination parameters"})
		return
	}
	orders, err := h.userOrdersService.GetUserOrders(c.Request.Context(), user.ID, params.Limit, params.Offset)
	if err != nil {
		h.logger.Errorw("Failed to get user orders", "requestID", requestID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if len(orders) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, orders)
}
