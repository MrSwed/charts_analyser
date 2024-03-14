package error

import (
	"errors"
	"github.com/lib/pq"
)

const (
	ErrUniqueViolationErr = pq.ErrorCode("23505")
)

var (
	ErrNotExist           = errors.New("not exist")
	ErrNotControlled      = errors.New("not controlled")
	ErrLocationOutOfRange = errors.New("location out of range")
)
