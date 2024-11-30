package api

import (
	"net/http"

	"cricket/cmd/app"
	"cricket/internal"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type AppInstance struct {
	App *app.App
}

func (service AppInstance) TournamentStats(c echo.Context) error {

	statsResponse, err := internal.QueryTournamentStats(service.App)
	if err != nil {
		errorResponse := map[string]string{"error": "Error in fetching stats!! Contact Admin"}
		service.App.Logger.Info("error in fetching stats", zap.Error(err))
		return c.JSON(http.StatusBadRequest, errorResponse)
	}
	return c.JSON(http.StatusOK, statsResponse)
}
