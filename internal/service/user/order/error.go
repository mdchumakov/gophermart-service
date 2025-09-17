package order

import "errors"

var (
	ErrOrderAlreadyExistsForUser = errors.New("order already exists for user")
	ErrOrderAlreadyProcessed     = errors.New("order already processed")
	ErrFailedToAddOrder          = errors.New("failed to add order")
	ErrBadOrderNumber            = errors.New("bad order number")
	ErrTooManyOrders             = errors.New("too many orders")
	ErrNoOrders                  = errors.New("no orders found")
)

func IsErrOrderAlreadyExistsForUser(err error) bool {
	return errors.Is(err, ErrOrderAlreadyExistsForUser)
}
func IsErrOrderAlreadyProcessedByAnotherUser(err error) bool {
	return errors.Is(err, ErrOrderAlreadyProcessed)
}
func IsErrFailedToAddOrder(err error) bool { return errors.Is(err, ErrFailedToAddOrder) }
func IsErrBadOrderNumber(err error) bool   { return errors.Is(err, ErrBadOrderNumber) }
