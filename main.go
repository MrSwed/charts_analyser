package main

//go:generate swag init --parseDependency --parseInternal

import (
	_ "charts_analyser/docs"
	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
	"log"
	"os"
	"strings"
)

const (
	EnvNameSwaggerAddress = "SWAGGER_ADDRESS"
	EnvNameSwaggerPort    = "SWAGGER_PORT"

	SwaggerPort    = "8000"
	SwaggerAddress = ":" + SwaggerPort
)

// @title Charts analyser: web-service API
// @version 1.0
// @host		localhost:3000

// @securityDefinitions.apikey BearerAuth
// @in             header
// @name           Authorization
// @description    Insert your access token default (Bearer access_token_here)

// @BasePath /api/

func main() {
	app := fiber.New()
	var (
		appAddr = SwaggerAddress
		appPort = SwaggerPort
	)
	if e := os.Getenv(EnvNameSwaggerAddress); e != "" {
		appAddr = e
	}
	if e := os.Getenv(EnvNameSwaggerPort); e != "" {
		appPort = e
	}

	checkAddr := strings.Split(appAddr, ":")
	if len(checkAddr) == 1 {
		appAddr = checkAddr[0] + ":" + appPort
	}

	app.Get("*", fiberSwagger.WrapHandler)

	err := app.Listen(appAddr)
	if err != nil {
		log.Fatalf("fiber.Listen failed %s", err)
	}
}
