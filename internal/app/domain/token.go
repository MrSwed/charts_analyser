package domain

import (
	"charts_analyser/internal/app/config"
	"charts_analyser/internal/app/constant"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

func NewClaimVessels(conf *config.JWT, id VesselID, name VesselName) *ClaimsAuth {
	return &ClaimsAuth{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   id.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(conf.TokenVesselLifeTime) * time.Second)),
		},
		Name: name.String(),
		Role: constant.RoleVessel,
		key:  conf.JWTSigningKey,
	}
}

func NewClaimOperator(conf *config.JWT, id UserID, name UserLogin) *ClaimsAuth {
	return &ClaimsAuth{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   id.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(conf.TokenLifeTime) * time.Second)),
		},
		Name: name.String(),
		Role: constant.RoleOperator,
		key:  conf.JWTSigningKey,
	}
}

func NewClaimAdmin(conf *config.JWT, id UserID, name UserLogin) *ClaimsAuth {
	c := NewClaimOperator(conf, id, name)
	c.Role = constant.RoleAdmin
	return c
}

type ClaimsAuth struct {
	jwt.RegisteredClaims
	Name string        `json:"name"`
	Role constant.Role `json:"role"`
	key  string
}

func (c *ClaimsAuth) Token() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, c)

	tokenString, err := token.SignedString([]byte(c.key))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
