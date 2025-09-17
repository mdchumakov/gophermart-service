package withdraw

import "errors"

var (
	ErrNotEnoughBalance = errors.New("not enough balance")
)

func IsErrNotEnoughBalance(err error) bool { return errors.Is(err, ErrNotEnoughBalance) }
