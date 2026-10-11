package main

import (
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/UTDNebula/nebula-platform-backend/internal/config"
	"github.com/UTDNebula/nebula-platform-backend/internal/controllers"
	"github.com/UTDNebula/nebula-platform-backend/internal/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db, err := config.NewPostgresPool()
	if err != nil {
		log.Fatalf("Connect to PostgreSQL: %v", err)
	}
	defer db.Close()
	log.Println("Connected to PostgreSQL")

	googleSettings, err := config.NewGoogleSettings()
	if err != nil {
		log.Fatalf("Connect to Google OIDC: %v", err)
	}

	google, err := config.NewOIDCClient(*googleSettings)
	if err != nil {
		log.Fatalf("Connect to Google OIDC: %v", err)
	}

	sessionManager, err := config.NewSessionManager(db)
	if err != nil {
		log.Fatalf("Initialize session manager: %v", err)
	}

	authController := controllers.NewAuthController(google, sessionManager, db)
	userController := controllers.NewUserController(db)

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

	// Register the routes
	routes.RegisterAuth(router, authController)
	routes.RegisterUser(router, userController, sessionManager)

	portString := config.GetPortString()
	log.Printf("Starting server on %s", portString)

	// Run the server
	err = http.ListenAndServe(
		portString,
		sessionManager.LoadAndSave(router),
	)
	if err != nil {
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
