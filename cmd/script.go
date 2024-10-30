package main

import (
	"cricket/cmd/app"
	"cricket/internal"
)

func main() {

	service := app.InitializeApp()
	service.Logger.Info("app initiated")

	internal.ReadData(service)
}
