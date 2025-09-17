package orders

import (
	"fmt"
	"gophermart-service/internal/base"
	"gophermart-service/internal/config"
	"gophermart-service/internal/service/jwt"
	serviceUserOrders "gophermart-service/internal/service/user/order"
	"io"
	"net/http"
	"strings"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

const maxBodySize int64 = 1024 * 1024 // 1 MB

type postUserOrdersHandler struct {
	logger            config.LoggerInterface
	userOrdersService serviceUserOrders.ServiceInterface
}

func NewPostUserOrdersHandler(
	logger config.LoggerInterface,
	userOrdersService serviceUserOrders.ServiceInterface,
) base.HandlerInterface {
	return &postUserOrdersHandler{
		logger:            logger,
		userOrdersService: userOrdersService,
	}
}

func (h *postUserOrdersHandler) Handle(c *gin.Context) {
	requestID := requestid.Get(c)

	user := jwt.ExtractUserFromContext(c.Request.Context())
	if user == nil {
		h.logger.Warnw("user not found in context", "requestID", requestID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	h.logger.Infow("Starting to handle user orders", "requestID", requestID)
	orderNumber, err := h.extractOrderNumber(c)
	if err != nil {
		return // Ошибка уже обработана в extractOrderNumber
	}

	if err = h.userOrdersService.LoadNewOrderNumber(c.Request.Context(), user.ID, orderNumber); err != nil {
		if serviceUserOrders.IsErrBadOrderNumber(err) {
			h.logger.Warnw("User order number is invalid", "requestID", requestID, "error", err)
			c.String(http.StatusUnprocessableEntity, err.Error())
			return
		}

		if serviceUserOrders.IsErrOrderAlreadyExistsForUser(err) {
			h.logger.Warnw(
				"Order already exists for user",
				"userID", user.ID,
				"orderNumber", orderNumber,
				"requestID", requestID)
			c.String(http.StatusOK, "Заказ уже загружен")
			return
		}

		if serviceUserOrders.IsErrOrderAlreadyProcessedByAnotherUser(err) {
			h.logger.Warnw("Order already processed by another user",
				"userID", user.ID,
				"orderNumber", orderNumber,
				"requestID", requestID)
			c.String(http.StatusConflict, "Заказ уже был загружен другим пользователем")
			return
		}

		if serviceUserOrders.IsErrFailedToAddOrder(err) {
			h.logger.Errorw("Failed to add order",
				"userID", user.ID,
				"orderNumber", orderNumber,
				"error", err,
				"requestID", requestID)
			c.String(
				http.StatusInternalServerError,
				fmt.Sprintf("Ошибка при добавлении заказа: %v", err.Error()),
			)
			return
		}
		h.logger.Errorw("Unexpected error when adding order",
			"userID", user.ID,
			"orderNumber", orderNumber,
			"error", err,
			"requestID", requestID)
		c.String(
			http.StatusInternalServerError,
			fmt.Sprintf("Неожиданная ошибка: %v", err.Error()),
		)
		return
	}

	h.logger.Infow("Order successfully processed",
		"userID", user.ID,
		"orderNumber", orderNumber,
		"requestID", requestID,
	)
	c.String(http.StatusAccepted, "Новый номер заказа принят в обработку")
}

func (h *postUserOrdersHandler) extractOrderNumber(c *gin.Context) (string, error) {
	requestID := requestid.Get(c)

	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBodySize))
	if err != nil {
		h.logger.Errorw("Failed to read request body", "body", string(body), "error", err, "requestID", requestID)
		c.String(http.StatusBadRequest, "Ошибка чтения данных")
		return "", err
	}

	orderNumber := strings.TrimSpace(string(body))
	if len(orderNumber) == 0 {
		h.logger.Warnw("Received empty orderNumber", "request_id", requestID)
		c.String(http.StatusBadRequest, "Пустой номер заказа")
		return "", err
	}

	return orderNumber, nil
}
