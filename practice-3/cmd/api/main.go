// @title Practice-3 API
// @version 1.0
// @description Users CRUD API
// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-KEY
package main

import (
	"log"

	"practice-3/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("app error: %v", err)
	}
}
