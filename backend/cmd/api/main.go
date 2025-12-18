package main

import (
	"log"

	drawhandler "backend/internal/adapter/http/handler"
	"backend/internal/app"
)

func main() {
	container, err := app.NewContainer()
	if err != nil {
		log.Fatalf("failed to initialize dependencies: %v", err)
	}

	router := drawhandler.NewRouter(container.DrawHandler)

	if err := router.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
