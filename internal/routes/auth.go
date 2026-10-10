package routes

import (
	"github.com/UTDNebula/nebula-platform-backend/internal/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterAuth(router *gin.Engine, controller *controllers.AuthController) {
	googleAuthGroup := router.Group("/api/auth")

	googleAuthGroup.GET("/google/login", controller.GoogleLogin)
	googleAuthGroup.GET("/google/callback", controller.GoogleCallback)
	googleAuthGroup.GET("/google/logout", controller.GoogleLogout)
}
