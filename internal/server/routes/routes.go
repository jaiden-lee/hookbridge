package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jaiden-lee/hookbridge/internal/server/handlers"
)

func NewRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.SetTrustedProxies(nil)

	router.GET("/", handlers.RootHandler)

	return router
}
