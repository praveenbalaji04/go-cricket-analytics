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

func (service AppInstance) ListOfTournaments(c echo.Context) error {
	var fields string
	err := internal.QueryTournamentsList(&fields, service.App.DB)
	if err != nil {
		service.App.Logger.Info("error in getting list of tournaments", zap.Error(err))
		var response = map[string]string{"error": "error in getting tournament list"}
		return c.JSON(http.StatusBadRequest, response)
	}
	var response = map[string]string{"match_types": fields}
	return c.JSON(http.StatusOK, response)
}
