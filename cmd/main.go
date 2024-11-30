package main

import (
	"net/http"

	"cricket/cmd/app"
	"cricket/service/router"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"go.uber.org/zap"
)

func main() {
	service := app.InitializeApp()
	service.Logger.Info("app initiated")

	e := echo.New()
	e.GET("/health", healthCheck)
	e.Logger.SetLevel(log.INFO)
	AddRouter(e, service)

	err := e.Start(":1323")
	if err != nil {
		service.Logger.Fatal("error in initiating server", zap.String("error", err.Error()))
	}
}

func healthCheck(c echo.Context) error {
	response := map[string]string{"status": "Hello world"}
	return c.JSON(http.StatusOK, response)
}

func AddRouter(e *echo.Echo, service *app.App) {
	v1 := e.Group("/v1/cricket")

	tournamentRouter := v1.Group("/tournament")
	router.AddTournamentRouters(tournamentRouter, service)
}
