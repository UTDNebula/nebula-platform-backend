package routes

import (
	"github.com/UTDNebula/nebula-platform-backend/internal/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterAuth(router *gin.Engine, controller *controllers.AuthController) {
	googleAuthGroup := router.Group("/api/auth")

	googleAuthGroup.GET("/:provider/login", controller.Login)
	googleAuthGroup.GET("/:provider/callback", controller.Callback)
	googleAuthGroup.GET("/:provider/logout", controller.Logout)
}
