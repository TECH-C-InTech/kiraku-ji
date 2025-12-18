package main

import (
	"log"

	"backend/internal/app"

	"github.com/gin-gonic/gin"
)

func main() {
	container, err := app.NewContainer()
	if err != nil {
		log.Fatalf("failed to initialize dependencies: %v", err)
	}

	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	log.Printf("draw fortune usecase initialized: %T", container.DrawFortuneUsecase)

	if err := router.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
