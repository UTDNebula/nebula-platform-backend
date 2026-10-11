package controllers

import (
	"crypto/rand"
	"crypto/subtle"
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

func Respond[T any](c *gin.Context, status int, message string, data T) {
	c.JSON(
		status,
		APIResponse[T]{
			Message: message,
			Data:    data,
		},
	)
}

func RespondInternalError(c *gin.Context, err error) {
	// TODO: Log these in Sentry
	log.Output(2, fmt.Sprintf("INTERNAL SERVER ERROR: %s", err.Error()))
	Respond(c, http.StatusInternalServerError, "error", err.Error())
}

func redirect(c *gin.Context, status int, url string) {
	c.Header("Referrer-Policy", "no-referrer")
	c.Header("Cache-Control", "no-store")
	c.Redirect(status, url)
}

func redirectError(c *gin.Context, err error) {
	// NOTE: We should log these in Sentry
	log.Output(2, fmt.Sprintf("AUTHENTICATION ERROR: %s", err.Error()))
	redirect(c, http.StatusSeeOther, "http://localhost:3000/login/error")
}

// equal does constant time compare for secure tokens
func equal(a string, b string) bool {
	byteA := []byte(a)
	byteB := []byte(b)
	if len(byteA) == 0 || len(byteB) == 0 || len(byteA) != len(byteB) {
		return false
	}
	return subtle.ConstantTimeCompare(byteA, byteB) == 1
}

// generateToken generate a randome 32-byte token, and Base64 encodes it to be safe for cookies and URLs
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
