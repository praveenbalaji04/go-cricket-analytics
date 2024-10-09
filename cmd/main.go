package main

import (
	"cricket/cmd/app"
)

func main() {

	service := app.InitializeApp()
	service.Logger.Info("app initiated")

	//internal.ReadData(service)
}
