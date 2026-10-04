package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/UTDNebula/nebula-platform-backend/internal/config"
	"github.com/alexedwards/scs/v2"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type AuthController struct {
	googleClient   *config.OIDCClient
	sessionManager *scs.SessionManager
}

// Initialize the auth controllers with OIDC client and session manager
func NewAuthController(
	googleClient *config.OIDCClient, sessionManager *scs.SessionManager,
) *AuthController {
	return &AuthController{
		googleClient:   googleClient,
		sessionManager: sessionManager,
	}
}

func (a *AuthController) Login(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// state verifies if the callback is for this login attempt
	state, err := generateToken()
	if err != nil {
		respondWithInternalError(c, fmt.Errorf("Unable to create state: %w", err))
		return
	}

	// PKCE verifier for exchanging authentication code from Google
	// for tokens
	pckeVerifier := oauth2.GenerateVerifier()

	// nonce verifies if the ID token is for this login attempt
	nonce, err := generateToken()
	if err != nil {
		respondWithInternalError(c, fmt.Errorf("Unable to create nonce: %w", err))
		return
	}

	// Session manager injects these secrets to request's context
	// When the request completes, scs will save the session to database
	a.sessionManager.Put(ctx, "google_state", state)
	a.sessionManager.Put(ctx, "google_nonce", nonce)
	a.sessionManager.Put(ctx, "google_pkce_verifier", pckeVerifier)

	// Build the URL to Google's login page with secrets that will be returned in the callback
	loginURL := a.googleClient.OAuthConfig.AuthCodeURL(
		state,
		oauth2.S256ChallengeOption(pckeVerifier),
		oidc.Nonce(nonce),
	)

	// The browser redirects to Google's login page
	c.Redirect(http.StatusFound, loginURL)
}
