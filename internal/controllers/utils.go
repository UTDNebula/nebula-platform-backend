package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse[T any] struct {
	Message string
	Data    T
}

func respond[T any](c *gin.Context, status int, message string, data T) {
	c.JSON(
		status,
		APIResponse[T]{
			Message: message,
			Data:    data,
		},
	)
}

// Helper function for logging and responding to a generic internal server error.
func respondInternalError(c *gin.Context, err error) {
	log.Output(2, fmt.Sprintf("INTERNAL SERVER ERROR: %s", err.Error()))
	respond(c, http.StatusInternalServerError, "error", err.Error())
}

// generateToken generate a randome 32-byte token, and Base64 encodes it to be safe for cookies and URLs
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
