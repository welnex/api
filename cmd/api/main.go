package main

import (
	_ "api/docs"
	"api/internal/config"
	"api/internal/pkg/app"
)

// @title       yodreik API
// @version     0.1
// @description API server for yodreik application
// @host        localhost:6969
// @BasePath    /api
//
// @securityDefinitions.apikey AccessToken
// @in                         header
// @name                       Authorization
func main() {
	c := config.MustLoad()
	a := app.New(c)

	a.Run()
}
