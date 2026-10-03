package main

import (
	"context"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/UTDNebula/nebula-platform-backend/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db, err := config.ConnectPostgres()
	if err != nil {
		log.Fatalf("connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	log.Println("Connected to PostgreSQL")

	router := gin.Default()

	// Enable CORS
	router.Use(cors.New(cors.Config{
		AllowOriginFunc:  isOriginAllowed,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "sentry-trace", "baggage"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           6 * time.Hour,
	}))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// Test the database connection
	router.GET("/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unavailable",
				"error":  "database is unavailable",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	})

	port := config.GetPortString()
	log.Printf("Starting server on %s", port)

	if err := router.Run(port); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}
}

// Allows only certain origins since this is a specific backend server
// instead of public api
func isOriginAllowed(origin string) bool {
	if gin.Mode() == gin.DebugMode {
		// Allows "*" if this is in development
		return true
	}

	// Deployment Vercel URL
	pattern := `^https?://nebula-platform(?:-[a-zA-Z0-9-]+)?\.vercel\.app/?$`
	regex := regexp.MustCompile(pattern)

	return origin == "http://localhost:3000" || regex.MatchString(origin)
}
