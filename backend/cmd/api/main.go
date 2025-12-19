package main

import (
	"context"
	"log"

	drawhandler "backend/internal/adapter/http/handler"
	"backend/internal/app"
	"backend/internal/config"
)

func main() {
	config.LoadDotEnv()

	ctx := context.Background()
	container, err := app.NewContainer(ctx)
	if err != nil {
		log.Fatalf("failed to initialize dependencies: %v", err)
	}
	defer func() {
		if closeErr := container.Close(); closeErr != nil {
			log.Printf("failed to close dependencies: %v", closeErr)
		}
	}()

	router := drawhandler.NewRouter(container.DrawHandler, container.PostHandler)

	if err := router.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
