package handler

import (
	"charts_analyser/internal/app/config"
	"charts_analyser/internal/app/constant"
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v2"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
)

func GetAccessWare(c *config.JWT) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey:    []byte(c.JWTSigningKey),
		SigningMethod: jwt.SigningMethodHS512.Name,
		TokenLookup:   "header:" + fiber.HeaderAuthorization,
		AuthScheme:    "Bearer",
		ContextKey:    constant.CtxStorageKey,
	})
}

func CheckIsRole(expRole constant.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !role(GetTokenClaims(c)).CheckIsRole(expRole) {
			_, err := c.Status(http.StatusForbidden).WriteString("wrong role")
			return err
		}
		return c.Next()
	}
}

func GetTokenClaims(c *fiber.Ctx) (claims jwt.MapClaims) {
	if u := c.Locals(constant.CtxStorageKey); u != nil {
		claims = u.(*jwt.Token).Claims.(jwt.MapClaims)
	}
	if claims == nil {
		claims = jwt.MapClaims{}
	}
	return
}

func role(claims jwt.MapClaims) constant.Role {
	return constant.Role(claims["role"].(float64))
}
