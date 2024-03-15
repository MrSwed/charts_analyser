package domain

import (
	"charts_analyser/internal/app/constant"
	"github.com/golang-jwt/jwt/v4"
)

func NewClaimVessels(id VesselID, name VesselName) *ClaimsAuth {
	return &ClaimsAuth{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: id.String(),
		},
		Name: name.String(),
		Role: constant.RoleVessel,
	}
}

func NewClaimOperator(id OperatorID, name OperatorName) *ClaimsAuth {
	return &ClaimsAuth{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: id.String(),
		},
		Name: name.String(),
		Role: constant.RoleOperator,
	}
}

type ClaimsAuth struct {
	jwt.RegisteredClaims
	Name string        `json:"name"`
	Role constant.Role `json:"role"`
}

func (c *ClaimsAuth) Token(key string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, c)

	tokenString, err := token.SignedString([]byte(key))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
