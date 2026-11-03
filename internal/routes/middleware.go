package routes

import (
	"net/http"
	"time"

	"github.com/UTDNebula/nebula-platform-backend/internal/controllers"
	"github.com/alexedwards/scs/v2"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func authMiddleware(manager *scs.SessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := manager.GetString(c.Request.Context(), "user_id")
		if userID == "" {
			respondUnauthorized(c)
			return
		}
		// To refresh the session
		manager.Put(c.Request.Context(), "last_active", time.Now().Unix())

		c.Set("user_id", userID)
		c.Next()
	}
}

func adminMiddleware(manager *scs.SessionManager, db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := manager.GetString(c.Request.Context(), "user_id")
		if userID == "" {
			respondUnauthorized(c)
			return
		}

		var role string
		if err := db.QueryRow(
			c.Request.Context(),
			"SELECT role FROM users WHERE id = $1",
			userID,
		).Scan(&role); err != nil {
			controllers.RespondInternalError(c, err)
			return
		}
		if role != "admin" {
			respondUnauthorized(c)
			return
		}

		// To refresh the session
		manager.Put(c.Request.Context(), "last_active", time.Now().Unix())

		c.Set("user_id", userID)
		c.Next()
	}
}

func respondUnauthorized(c *gin.Context) {
	controllers.Respond(c, http.StatusUnauthorized, "error", "Unauthorized")
}
