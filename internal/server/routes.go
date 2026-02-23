package server

import "github.com/gin-gonic/gin"

func GetRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/", pingHandler)

	router.POST("/api/connect", connectHandlers.connectOrCreateTunnel)

	router.Any("/tunnel/:tunnel_name/*proxyPath", tunnelHandlers.handleHttpRequestForTunnel)

	return router
}

func pingHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Hello World!",
	})
}
