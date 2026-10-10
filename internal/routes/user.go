package routes

import (
	"github.com/UTDNebula/nebula-platform-backend/internal/controllers"
	"github.com/alexedwards/scs/v2"
	"github.com/gin-gonic/gin"
)

func RegisterUser(router *gin.Engine, controller *controllers.UserController, manager *scs.SessionManager) {
	userGroup := router.Group("/api/users")
	userGroup.Use(authMiddleware(manager))

	userGroup.GET("/profile", controller.Profile)
}
