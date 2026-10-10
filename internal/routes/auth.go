package routes

import (
	"github.com/UTDNebula/nebula-platform-backend/internal/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterAuth(router *gin.Engine, controller *controllers.AuthController) {
	googleAuthGroup := router.Group("/api/auth/google")

	googleAuthGroup.GET("/login", controller.Login)
	googleAuthGroup.GET("/callback", controller.Callback)
	googleAuthGroup.GET("/logout", controller.Logout)
}
