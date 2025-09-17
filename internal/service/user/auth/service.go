package auth

import (
	"context"
	"gophermart-service/internal/base"
	"gophermart-service/internal/config"
	usersRepo "gophermart-service/internal/repository/users"

	"golang.org/x/crypto/bcrypt"
)

const PasswordHashCost = 15

// Service представляет сервис регистрации пользователей
type Service struct {
	logger config.LoggerInterface
	repo   usersRepo.RepositoryInterface
}

// NewRegisterService создает новый экземпляр сервиса регистрации
func NewRegisterService(
	logger config.LoggerInterface,
	repo usersRepo.RepositoryInterface,
) ServiceInterface {
	return &Service{
		logger: logger,
		repo:   repo,
	}
}

// RegisterUser выполняет регистрацию нового пользователя
func (s *Service) RegisterUser(ctx context.Context, dtoIn *InDTO) (*OutDTO, error) {
	requestID := base.GetRequestID(ctx)

	s.logger.Infow("User registration initiated",
		"requestID", requestID,
		"login", dtoIn.Login)

	// Валидация входных данных
	if err := dtoIn.validateRequest(); err != nil {
		s.logger.Warnf("Registration validation failed: requestID=%s, error=%s",
			requestID, err.Error())
		return nil, err
	}

	// Хеширование пароля
	passwordHash, err := s.HashPassword(dtoIn.Password)
	if err != nil {
		s.logger.Errorw("Password hashing failed",
			"requestID", requestID,
			"error", err.Error())
		return nil, ErrFiledToProcessPassword
	}

	// Сохранение пользователя в БД
	userID, err := s.repo.Add(ctx, dtoIn.Login, passwordHash)
	if err != nil {
		if usersRepo.IsErrUserLoginAlreadyExists(err) {
			return nil, ErrUserLoginAlreadyExists
		}
		s.logger.Errorw("User registration failed",
			"requestID", requestID,
			"login", dtoIn.Login,
			"error", err.Error(),
		)
		return nil, err
	}

	s.logger.Infow("User registration completed successfully",
		"requestID", requestID,
		"login", dtoIn.Login,
		"userID", userID)

	return &OutDTO{
		UserID: userID,
	}, nil
}

// LoginUser выполняет аутентификацию пользователя
func (s *Service) LoginUser(ctx context.Context, dtoIn *InDTO) (*OutDTO, error) {
	requestID := base.GetRequestID(ctx)

	s.logger.Infow("User login initiated",
		"requestID", requestID,
		"login", dtoIn.Login)

	// Валидация входных данных
	if err := dtoIn.validateRequest(); err != nil {
		s.logger.Warnf("Registration validation failed: requestID=%s, error=%s",
			requestID, err.Error())
		return nil, err
	}

	userID, hashPassword, err := s.repo.GetUserHashPassword(ctx, dtoIn.Login)
	if err != nil {
		if usersRepo.IsErrUserNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	err = s.VerifyPassword(hashPassword, dtoIn.Password)
	if err != nil {
		return nil, ErrBadPassword
	}

	return &OutDTO{UserID: userID}, nil
}

// HashPassword хеширует пароль с использованием bcrypt
func (s *Service) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), PasswordHashCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyPassword проверяет пароль против хеша
func (s *Service) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
