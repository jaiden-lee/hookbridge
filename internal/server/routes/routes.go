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
		tunnelRoutes.Any(":/project_id/*proxyPath", handlers.TunnelHandlers.WebhookForwardingHandler)
	}

	projectRoutes := router.Group("/api/projects")
	projectRoutes.Use(middleware.AuthMiddleware())
	{
		projectRoutes.POST("", handlers.ProjectHandlers.CreateProjectHandler)
		projectRoutes.GET("", handlers.ProjectHandlers.GetProjectsHandler)
		projectRoutes.PATCH("/:project_id/password", handlers.ProjectHandlers.ChangeProjectPasswordHandler)
		projectRoutes.DELETE("/:project_id", handlers.ProjectHandlers.DeleteProjectHandler)
	}

	connectionRoutes := router.Group("/api/connect")
	{
		connectionRoutes.POST("/:project_id", handlers.ConnectHandlers.ConnectToProjectHandler)
		connectionRoutes.GET("/:project_id/status", handlers.ConnectHandlers.GetConnectionStatusHandler)
	}

	router.GET("/", handlers.RootHandler)

	return router
}
