package server

import "github.com/gin-gonic/gin"

func GetRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/", pingHandler)

	router.POST("/api/connect", connectHandlers.connectOrCreateTunnel)

	return router
}

func pingHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hello World!",
	})
}
