package routes

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/gin-gonic/gin"
)

func authMiddleware(manager *scs.SessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := manager.GetString(c.Request.Context(), "user_id")
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			return
		}
		// To refresh the session
		manager.Put(c.Request.Context(), "last_active", time.Now().Unix())

		c.Set("user_id", userID)
		c.Next()
	}
}
