package domain

import (
	"github.com/golang-jwt/jwt/v4"
)

type ClaimsOperator struct {
	jwt.RegisteredClaims
	Name string `json:"name"`
	Role int64  `json:"role"`
}

func (c *ClaimsOperator) Token(key string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, ClaimsOperator{
		RegisteredClaims: jwt.RegisteredClaims{},
		Name:             "Simulator",
		Role:             2,
	})

	tokenString, err := token.SignedString([]byte(key))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
