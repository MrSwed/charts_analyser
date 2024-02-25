package error

import "errors"

var (
	ErrNotExist           = errors.New("not exist")
	ErrNotControlled      = errors.New("not controlled")
	ErrLocationOutOfRange = errors.New("location out of range")
)
