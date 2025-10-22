package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/jaiden-lee/hookbridge/internal/server/handlers"
	"github.com/jaiden-lee/hookbridge/internal/server/middleware"
	"github.com/jaiden-lee/hookbridge/internal/server/utils"
)

func NewRouter() *gin.Engine {
	utils.AuthService.Init()

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.SetTrustedProxies(nil)

	tunnelRoutes := router.Group("/tunnel")
	{
		tunnelRoutes.Any(":/project_id/*proxyPath")
	}

	projectRoutes := router.Group("/api/projects")
	projectRoutes.Use(middleware.AuthMiddleware())
	{
		projectRoutes.POST("")
		projectRoutes.GET("")
		projectRoutes.PATCH("/:project_id/password")
		projectRoutes.DELETE("/:project_id")
	}

	connectionRoutes := router.Group("/api/connect")
	{
		connectionRoutes.POST("/:project_id")
		connectionRoutes.GET("/:project_id/status")
	}

	router.GET("/", handlers.RootHandler)

	return router
}
