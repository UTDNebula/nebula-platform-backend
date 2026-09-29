package main

import (
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/UTDNebula/nebula-platform-backend/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Set up third party clients here
	if _, err := config.ConnectMongo(); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}

	router := gin.Default()

	// Enable CORS
	router.Use(cors.New(cors.Config{
		AllowOriginFunc:  isOriginAllowed,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "sentry-trace", "baggage"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           6 * time.Hour,
	}))

	// Health endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Connect routes here

	// Run the router
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
