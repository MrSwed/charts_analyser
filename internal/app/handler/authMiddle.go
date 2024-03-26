package handler

import (
	"charts_analyser/internal/app/config"
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v2"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
)

func GetAccessWare(confJWT *config.JWT) fiber.Handler {
	return jwtware.New(jwtware.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			_, err = c.Status(http.StatusUnauthorized).WriteString(err.Error())
			return err
		},
		SigningKey:    []byte(confJWT.JWTSigningKey),
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

func GetVesselID(c *fiber.Ctx) (id domain.VesselID) {
	claims := GetTokenClaims(c)
	if role(claims).CheckIsRole(constant.RoleVessel) {
		switch cl := claims[constant.TokenIDKey].(type) {
		case float64:
			id = domain.VesselID(cl)
		case string:
			_ = id.SetFromStr(cl)
		}
	}
	return
}

func GetTokenClaims(c *fiber.Ctx) (claims jwt.MapClaims) {
	if u := c.Locals(constant.CtxStorageKey); u != nil {
		if cl, ok := u.(*jwt.Token); ok {
			claims, _ = cl.Claims.(jwt.MapClaims)
		}
	}
	if claims == nil {
		claims = jwt.MapClaims{}
	}
	return
}

func role(claims jwt.MapClaims) constant.Role {
	return constant.Role(claims[constant.TokenRoleKey].(float64))
}
