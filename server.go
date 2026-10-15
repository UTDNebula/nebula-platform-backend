package main

import (
	"log"
	"net/http"
	"regexp"
	"time"

	_ "github.com/UTDNebula/nebula-platform-backend/docs"
	"github.com/UTDNebula/nebula-platform-backend/internal/config"
	"github.com/UTDNebula/nebula-platform-backend/internal/controllers"
	"github.com/UTDNebula/nebula-platform-backend/internal/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/oauth2"
)

//go:generate go tool swag init -g server.go --parseInternal --propertyStrategy pascalcase --exclude test-frontend -o docs

// @title Nebula Platform Backend API
// @version 1.0
// @description Authentication and user endpoints for the Nebula Platform backend. Login and callback use browser redirects and a server-side session identified by the nebula_session cookie.
// @BasePath /
// @securityDefinitions.apikey SessionAuth
// @in header
// @name Cookie
// @description Browser-managed nebula_session cookie. Log in using the same browser. Swagger 2.0 represents cookie authentication as a header; browsers send cookies automatically and do not allow JavaScript to set the Cookie header manually.
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

	googleProvider, err := config.NewOIDCAuthProvider(
		*googleSettings,
		[]oauth2.AuthCodeOption{oauth2.SetAuthURLParam("prompt", "select_account")},
		true,
		"google",
	)
	if err != nil {
		log.Fatalf("Connect to Google OIDC: %v", err)
	}

	// TODO: Add new auth provider here to the map

	authProviders := map[string]config.AuthProvider{
		"google": *googleProvider,
	}

	sessionManager, err := config.NewSessionManager(db)
	if err != nil {
		log.Fatalf("Initialize session manager: %v", err)
	}

	authController := controllers.NewAuthController(authProviders, sessionManager, db)
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

	// Serve generated API documentation.
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
