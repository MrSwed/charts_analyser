package domain

import (
	"charts_analyser/internal/app/constant"
	"strconv"
	"time"
)

type User struct {
	ID    UserID    `json:"id" db:"id"`
	Login UserLogin `json:"name" db:"name"`
}

type UserID int64

func (v *UserID) String() string {
	return strconv.FormatInt(int64(*v), 10)
}

type UserLogin string

func (v *UserLogin) String() string {
	return string(*v)
}

type UserDB struct {
	User
	CreatedAt  time.Time     `db:"created_at" json:"createdAt"`
	ModifiedAt time.Time     `db:"modified_at" json:"modifiedAt"`
	Hash       []byte        `db:"hash" json:"-"`
	Role       constant.Role `db:"role" json:"role"`
	IsDeleted  bool          `db:"is_deleted" json:"-"`
}

// Input

type Password string

func (p Password) String() string {
	return string(p)
}

type Auth struct {
	Login    UserLogin `json:"login" validate:"required,min=6"`
	Password Password  `json:"password" validate:"required,password"`
}

type UserChange struct {
	Auth
	Role constant.Role `json:"role" validate:"required,oneof=2 4"`
	ID   *UserID       `json:"id,omitempty" validate:"omitempty"`
}
