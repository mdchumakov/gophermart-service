package users

import "errors"

var (
	ErrUserLoginAlreadyExists = errors.New("user login already exists")
	ErrUserNotFound           = errors.New("user not found")
)

func IsErrUserLoginAlreadyExists(err error) bool {
	return errors.Is(err, ErrUserLoginAlreadyExists)
}

func IsErrUserNotFound(err error) bool {
	return errors.Is(err, ErrUserNotFound)
}
