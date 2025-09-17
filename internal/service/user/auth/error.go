package auth

import "errors"

var (
	ErrLoginIsRequired        = errors.New("login is required")
	ErrPasswordIsRequired     = errors.New("password is required")
	ErrPasswordTooShort       = errors.New("password is too short")
	ErrLoginTooShort          = errors.New("login is too short")
	ErrLoginTooLong           = errors.New("login is too long")
	ErrPasswordTooLong        = errors.New("password is too long")
	ErrFiledToProcessPassword = errors.New("filed to process password")
	ErrUserLoginAlreadyExists = errors.New("user login already exists")
	ErrBadPassword            = errors.New("bad password")
	ErrUserNotFound           = errors.New("user not found")
)

func IsErrLoginIsRequired(err error) bool {
	return errors.Is(err, ErrLoginIsRequired)
}

func IsErrPasswordIsRequired(err error) bool {
	return errors.Is(err, ErrPasswordIsRequired)
}

func IsErrPasswordTooShort(err error) bool {
	return errors.Is(err, ErrPasswordTooShort)
}

func IsErrLoginTooShort(err error) bool {
	return errors.Is(err, ErrLoginTooShort)
}

func IsErrLoginTooLong(err error) bool {
	return errors.Is(err, ErrLoginTooLong)
}

func IsErrPasswordTooLong(err error) bool {
	return errors.Is(err, ErrPasswordTooLong)
}

func IsErrUserLoginAlreadyExists(err error) bool {
	return errors.Is(err, ErrUserLoginAlreadyExists)
}

func IsBadPassword(err error) bool {
	return errors.Is(err, ErrBadPassword)
}

func IsErrUserNotFound(err error) bool { return errors.Is(err, ErrUserNotFound) }
