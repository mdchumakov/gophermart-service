package auth

import "strings"

const (
	MaxLoginLength    = 50
	MaxPasswordLength = 100
	MinLoginLength    = 3
	MinPasswordLength = 6
)

// InDTO представляет запрос на регистрацию пользователя
type InDTO struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// validateRequest выполняет валидацию запроса регистрации
func (in *InDTO) validateRequest() error {
	// Проверка на пустые поля
	if strings.TrimSpace(in.Login) == "" {
		return ErrLoginIsRequired
	}
	if strings.TrimSpace(in.Password) == "" {
		return ErrPasswordIsRequired
	}

	// Проверка минимальной длины логина
	if len(in.Login) < MinLoginLength {
		return ErrLoginTooShort
	}

	// Проверка минимальной длины пароля
	if len(in.Password) < MinPasswordLength {
		return ErrPasswordTooShort
	}

	// Проверка максимальной длины (защита от DoS атак)
	if len(in.Login) > MaxLoginLength {
		return ErrLoginTooLong
	}
	if len(in.Password) > MaxPasswordLength {
		return ErrPasswordTooLong
	}

	return nil
}

// OutDTO представляет ответ при успешной регистрации
type OutDTO struct {
	UserID int `json:"user_id"`
}
