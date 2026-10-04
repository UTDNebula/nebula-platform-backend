package routes

import (
	"github.com/UTDNebula/nebula-platform-backend/internal/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterAuth(router *gin.Engine, controller *controllers.AuthController) {
	googleAuthGroup := router.Group("/auth/google")

	googleAuthGroup.GET("login", controller.Login)
	googleAuthGroup.GET("callback", func(ctx *gin.Context) {})
	googleAuthGroup.GET("logout", func(ctx *gin.Context) {})
}
