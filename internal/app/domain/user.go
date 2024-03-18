package domain

import (
	"charts_analyser/internal/app/constant"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"time"
)

type UserID int64

func (v *UserID) String() string {
	return strconv.FormatInt(int64(*v), 10)
}

type UserLogin string

func (v *UserLogin) String() string {
	return string(*v)
}

type UserDB struct {
	ID         UserID        `db:"id" json:"id" `
	Login      UserLogin     `db:"login" json:"login" `
	CreatedAt  time.Time     `db:"created_at" json:"createdAt"`
	ModifiedAt time.Time     `db:"modified_at" json:"modifiedAt"`
	Hash       Hash          `db:"hash" json:"-"`
	Role       constant.Role `db:"role" json:"role"`
	IsDeleted  bool          `db:"is_deleted" json:"-"`
}

func NewUserDB(id UserID, login UserLogin, passwd *Password, role constant.Role) (u *UserDB, err error) {
	var hash []byte
	if passwd != nil {
		hash, err = passwd.Hash()
		if err != nil {
			return
		}
	}
	u = &UserDB{
		ID:         id,
		Login:      login,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
		Role:       role,
	}
	if hash != nil {
		u.Hash = hash
	}
	return
}

type Hash []byte

func (h Hash) IsValidPassword(p Password) bool {
	err := bcrypt.CompareHashAndPassword(h, []byte(p))
	return err == nil
}

// Input

type Password string

func (p Password) String() string {
	return string(p)
}

func (p *Password) Hash() ([]byte, error) {
	if p == nil {
		return nil, nil
	}
	return bcrypt.GenerateFromPassword([]byte(*p), bcrypt.DefaultCost)
}

type LoginForm struct {
	Login    UserLogin `json:"login" validate:"required"`
	Password Password  `json:"password" validate:"required"`
}

type UserNewForm struct {
	Login    UserLogin `json:"login" validate:"required,min=6"`
	Password Password  `json:"password" validate:"required,password"`
}

type UserChange struct {
	Login    UserLogin     `json:"login" validate:"required,min=6"`
	Password *Password     `json:"password,omitempty" validate:"omitempty,password"`
	Role     constant.Role `json:"role" validate:"required,oneof=2 4"`
	ID       *UserID       `json:"id,omitempty" validate:"omitempty"`
}
