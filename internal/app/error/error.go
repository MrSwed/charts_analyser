package error

import "errors"

var (
	ErrNotExist      = errors.New("not exist")
	ErrNotControlled = errors.New("not controlled")
)
