package base

import "fmt"

type Exception struct {
	Message   string
	Operation string
	Err       error
}

func (e *Exception) Error() string {
	return fmt.Sprintf("%s %s: %v", e.Message, e.Operation, e.Err)
}

func (e *Exception) Unwrap() error {
	return e.Err
}
