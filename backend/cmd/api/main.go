package main

import (
	"context"
	"fmt"
	"log"

	drawhandler "backend/internal/adapter/http/handler"
	"backend/internal/app"
	"backend/internal/config"
)

func main() {
	config.LoadDotEnv()

	if err := run(context.Background()); err != nil {
		log.Fatalf("API起動失敗: %v", err)
	}
}

func run(ctx context.Context) error {
	container, err := app.NewContainer(ctx)
	if err != nil {
		return fmt.Errorf("依存初期化失敗: %w", err)
	}
	defer func() {
		if closeErr := container.Close(); closeErr != nil {
			log.Printf("依存終了失敗: %v", closeErr)
		}
	}()

	router := drawhandler.NewRouter(container.DrawHandler, container.PostHandler)
	if err := router.Run(); err != nil {
		return fmt.Errorf("サーバー起動失敗: %w", err)
	}
	return nil
}
