package routes

import (
	"github.com/UTDNebula/nebula-platform-backend/internal/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterAuth(router *gin.Engine, controller *controllers.AuthController) {
	authGroup := router.Group("/api/auth")

	authGroup.GET("/:provider/login", controller.Login)
	authGroup.GET("/:provider/callback", controller.Callback)
	authGroup.GET("/:provider/logout", controller.Logout)
}
