package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// Enable CORS
	router.Use(CORS)

	router.GET("/health", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Enable Logging
	router.Use(LogRequest)

	// Connect routes

	// Get the port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("starting server on :%s", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func CORS(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, x-api-key, Origin, Content-type, Authorization, sentry-trace, baggage")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET")

	if c.Request.Method == "OPTIONS" {
		c.IndentedJSON(204, "")
		return
	}

	c.Next()
}

func LogRequest(c *gin.Context) {
	log.Printf("%s %s %s", c.Request.Method, c.Request.URL.Path, c.Request.Host)
	c.Next()
}
