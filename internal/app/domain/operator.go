package domain

import "strconv"

type Operator struct {
	ID   OperatorID   `json:"id"`
	Name OperatorName `json:"name"`
}

type OperatorID int64

func (v *OperatorID) String() string {
	return strconv.FormatInt(int64(*v), 10)
}

type OperatorName string

func (v *OperatorName) String() string {
	return string(*v)
}
