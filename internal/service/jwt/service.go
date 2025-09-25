package jwt

import (
	"context"
	"fmt"
	"gophermart-service/internal/base"
	"gophermart-service/internal/config"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims представляет структуру JWT claims
type Claims struct {
	UserID    int    `json:"user_id"`
	UserLogin string `json:"user_login"`
	jwt.RegisteredClaims
}

type jwtService struct {
	secretKey     string
	tokenDuration time.Duration
	issuer        string
	algorithm     string
	logger        config.LoggerInterface
}

// NewJWTService создает новый экземпляр JWT сервиса
func NewJWTService(settings *config.JWTSettings, logger config.LoggerInterface) ServiceInterface {
	return &jwtService{
		secretKey:     settings.SecretKey,
		tokenDuration: settings.TokenDuration,
		issuer:        settings.Issuer,
		algorithm:     settings.Algorithm,
		logger:        logger,
	}
}

// GenerateToken генерирует JWT токен для конкретного пользователя
func (s *jwtService) GenerateToken(ctx context.Context, user *InDTO) (string, error) {
	requestID := base.GetRequestID(ctx)

	if user == nil {
		s.logger.Errorw("User cannot be nil")
		return "", ErrUserCannotBeNil
	}

	// Создаем claims для токена
	now := time.Now()
	claims := Claims{
		UserID:    user.ID,
		UserLogin: user.Login,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   user.Login,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenDuration)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	// Создаем токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписываем токен
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		s.logger.Errorw(
			"Failed to sign JWT token",
			"user_id", user.ID,
			"username", user.Login,
			"error", err,
			"request_id", requestID,
		)
		return "", &ErrTokenSigning{
			Exception: base.Exception{Message: "Failed to sign JWT token", Operation: "JWT signing", Err: err},
		}
	}
	s.logger.Debugw("JWT token generated successfully",
		"user_id", user.ID,
		"username", user.Login,
		"expires_at", claims.ExpiresAt,
		"request_id", requestID,
	)
	return tokenString, nil
}

// ValidateToken валидирует JWT токен и возвращает информацию о пользователе
func (s *jwtService) ValidateToken(ctx context.Context, tokenString string) (*InDTO, error) {
	requestID := base.GetRequestID(ctx)

	if tokenString == "" {
		s.logger.Errorw("Token cannot be empty", "request_id", requestID)
		return nil, ErrTokenCannotBeEmpty
	}

	// Убираем префикс "Bearer" если он есть
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// Парсим токен
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			s.logger.Errorw("Unexpected signing method", "method", token.Header["alg"], "request_id", requestID)
			return nil, &ErrUnexpectedSigningMethod{
				Exception: base.Exception{
					Message:   fmt.Sprintf("unexpected signing method: %v", token.Header["alg"]),
					Operation: "JWT signing",
					Err:       nil},
			}
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		s.logger.Errorw("Failed to parse JWT token", "error", err, "request_id", requestID)
		return nil, &ErrFailedToParseToken{
			Exception: base.Exception{Message: "Failed to parse token", Operation: "JWT signing", Err: err},
		}
	}

	// Проверяем валидность токена
	if !token.Valid {
		s.logger.Errorw("Invalid token", "error", err, "request_id", requestID)
		return nil, ErrInvalidTokenFormat
	}

	// Извлекаем claims
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidTokenClaims
	}

	// Создаем модель пользователя
	user := &InDTO{
		ID:    claims.UserID,
		Login: claims.UserLogin,
	}

	s.logger.Debugw(
		"JWT token validated successfully",
		"user_id", user.ID,
		"user_login", user.Login,
		"request_id", requestID,
	)
	return user, nil
}

// GetUserLoginFromToken извлекает ID пользователя из JWT токена
func (s *jwtService) GetUserLoginFromToken(ctx context.Context, tokenString string) (string, error) {
	user, err := s.ValidateToken(ctx, tokenString)
	if err != nil {
		return "", err
	}

	return user.Login, nil
}

// RefreshToken обновляет JWT токен
func (s *jwtService) RefreshToken(ctx context.Context, tokenString string) (string, error) {
	// Валидируем текущий токен
	user, err := s.ValidateToken(ctx, tokenString)
	if err != nil {
		return "", err
	}

	// Генерируем новый токен
	return s.GenerateToken(ctx, user)
}

// IsTokenExpired проверяет, истек ли срок действия токена
func (s *jwtService) IsTokenExpired(ctx context.Context, tokenString string) (bool, error) {
	requestID := base.GetRequestID(ctx)

	if tokenString == "" {
		return true, ErrTokenCannotBeEmpty
	}

	// Убираем префикс "Bearer" если он есть
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// Парсим токен без валидации подписи для проверки срока действия
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return true, &ErrFailedToParseToken{
			Exception: base.Exception{Message: "Failed to parse token", Operation: "JWT signing", Err: err},
		}
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		s.logger.Errorw("Invalid token claims", "error", err, "request_id", requestID)
		return true, ErrInvalidTokenClaims
	}

	// Проверяем срок действия
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		s.logger.Debugw("Token expired", "request_id", requestID)
		return true, nil
	}

	return false, nil
}

// GetTokenExpirationTime возвращает время истечения токена
func (s *jwtService) GetTokenExpirationTime(ctx context.Context, tokenString string) (*time.Time, error) {
	requestID := base.GetRequestID(ctx)

	if tokenString == "" {
		s.logger.Errorw("Token cannot be empty", "request_id", requestID)
		return nil, ErrTokenCannotBeEmpty
	}

	// Убираем префикс "Bearer " если он есть
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// Парсим токен без валидации подписи
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		s.logger.Errorw("Failed to parse JWT token", "error", err, "request_id", requestID)
		return nil, &ErrFailedToParseToken{
			Exception: base.Exception{Message: "Failed to parse token", Operation: "JWT signing", Err: err},
		}
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		s.logger.Errorw("Invalid token claims", "error", err, "request_id", requestID)
		return nil, ErrInvalidTokenClaims
	}

	if claims.ExpiresAt == nil {
		s.logger.Errorw("Token has no expiration time", "request_id", requestID)
		return nil, ErrTokenHasNoExpirationTime
	}

	return &claims.ExpiresAt.Time, nil
}
