package handler

import "github.com/gin-gonic/gin"

// NewRouter は HTTP ハンドラーを紐づけた gin.Engine を返す。
func NewRouter(drawHandler *DrawHandler, postHandler *PostHandler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/draws/random", drawHandler.GetRandomDraw)
	router.POST("/posts", postHandler.CreatePost)

	return router
}
