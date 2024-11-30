package router

import (
	"cricket/cmd/app"
	"cricket/service/api"

	"github.com/labstack/echo/v4"
)

func AddTournamentRouters(e *echo.Group, service *app.App) {
	tournament := api.AppInstance{App: service}
	e.GET("/stats", tournament.TournamentStats)
}
