package register

import (
	"fmt"
	"gophermart-service/internal/base"
	"gophermart-service/internal/config"
	serviceJWT "gophermart-service/internal/service/jwt"
	serviceUserAuth "gophermart-service/internal/service/user/auth"
	"net/http"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

type postUserRegisterHandler struct {
	logger          config.LoggerInterface
	userAuthService serviceUserAuth.ServiceInterface
	jwtService      serviceJWT.ServiceInterface
	jwtSettings     *config.JWTSettings
}

func NewPostRegisterHandler(
	logger config.LoggerInterface,
	userAuthService serviceUserAuth.ServiceInterface,
	jwtService serviceJWT.ServiceInterface,
	jwtSettings *config.JWTSettings,
) base.HandlerInterface {
	return &postUserRegisterHandler{
		logger:          logger,
		userAuthService: userAuthService,
		jwtService:      jwtService,
		jwtSettings:     jwtSettings,
	}
}

// RequestBodyInDTO представляет запрос на регистрацию пользователя
type RequestBodyInDTO struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ResponseBodyOutDTO представляет ответ при успешной регистрации
type ResponseBodyOutDTO struct {
	AuthToken string `json:"x_auth_token"`
}

func (h *postUserRegisterHandler) Handle(c *gin.Context) {
	var dtoIn RequestBodyInDTO
	requestID := requestid.Get(c)
	h.logger.Infow("Starting user registration", "requestID", requestID)

	if err := h.parseRequestBody(c, &dtoIn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response, err := h.userAuthService.RegisterUser(
		c.Request.Context(),
		&serviceUserAuth.InDTO{
			Login:    dtoIn.Login,
			Password: dtoIn.Password,
		})

	if err != nil {
		if serviceUserAuth.IsErrLoginIsRequired(err) ||
			serviceUserAuth.IsErrPasswordIsRequired(err) ||
			serviceUserAuth.IsErrLoginTooShort(err) ||
			serviceUserAuth.IsErrPasswordTooShort(err) ||
			serviceUserAuth.IsErrLoginTooLong(err) ||
			serviceUserAuth.IsErrPasswordTooLong(err) {
			h.logger.Warnw("Validation error during user registration",
				"error", err,
				"request_id", requestID,
				"login", dtoIn.Login)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		} else if serviceUserAuth.IsErrUserLoginAlreadyExists(err) {
			h.logger.Warnw("User login already exists",
				"request_id", requestID,
				"login", dtoIn.Login)
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("User login '%s' already exists", dtoIn.Login)})
			return
		}
		h.logger.Errorw("Failed to register user",
			"error", err,
			"request_id", requestID,
			"login", dtoIn.Login)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	jwtToken, err := h.jwtService.GenerateToken(c.Request.Context(), &serviceJWT.InDTO{
		ID:    response.UserID,
		Login: dtoIn.Login,
	})
	if err != nil {
		h.logger.Errorw("Failed to generate JWT token",
			"error", err,
			"request_id", requestID,
			"login", dtoIn.Login)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate auth token"})
		return
	}

	base.SetTokenToCookie(c, jwtToken, h.jwtSettings, h.jwtSettings.TokenDuration)
	base.SetTokenToHeader(c, jwtToken)

	c.JSON(http.StatusOK, gin.H{
		"user_id": response.UserID,
		"message": "User registered successfully",
		"token":   jwtToken,
	})
}

func (h *postUserRegisterHandler) parseRequestBody(c *gin.Context, dtoIn *RequestBodyInDTO) error {
	requestID := requestid.Get(c)

	if err := c.ShouldBindJSON(dtoIn); err != nil {
		h.logger.Warnw("Invalid JSON in request body",
			"error", err,
			"request_id", requestID,
			"remote_addr", c.Request.RemoteAddr)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return err
	}

	return nil
}
