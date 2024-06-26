package error

import (
	"errors"
)

var (
	ErrNotExist           = errors.New("not exist")
	ErrNotControlled      = errors.New("not controlled")
	ErrLocationOutOfRange = errors.New("location out of range")
	ErrDuplicateRecord    = errors.New("duplicate record")
	ErrLogin              = errors.New("bad pair login/password")
)
