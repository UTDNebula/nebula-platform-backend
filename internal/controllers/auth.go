package controllers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/UTDNebula/nebula-platform-backend/internal/config"
	"github.com/alexedwards/scs/v2"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
)

const (
	googleProvider = "google"
)

type AuthController struct {
	GoogleClient *config.OIDCClient
	Manager      *scs.SessionManager
	DB           *pgxpool.Pool
}

// Initialize the auth controllers with OIDC client and session manager
func NewAuthController(
	googleClient *config.OIDCClient,
	sessionManager *scs.SessionManager,
	db *pgxpool.Pool,
) *AuthController {
	return &AuthController{
		GoogleClient: googleClient,
		Manager:      sessionManager,
		DB:           db,
	}
}

func (a *AuthController) GoogleLogin(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// state verifies if callback ties to this login
	state, err := generateToken()
	if err != nil {
		redirectError(c, fmt.Errorf("Unable to create state: %w", err))
		return
	}

	// PKCE verifier for exchanging authz code for oauth2 token
	pckeVerifier := oauth2.GenerateVerifier()

	// nonce verifies ID token
	nonce, err := generateToken()
	if err != nil {
		redirectError(c, fmt.Errorf("Unable to create nonce: %w", err))
		return
	}

	// Session manager injects these secrets to request's context
	// When the request completes, scs will save the session to database
	a.Manager.Put(ctx, "state", state)
	a.Manager.Put(ctx, "pkce_verifier", pckeVerifier)
	a.Manager.Put(ctx, "nonce", nonce)

	// Build the URL to Google's login page
	loginURL := a.GoogleClient.OAuthConfig.AuthCodeURL(
		state,
		oauth2.S256ChallengeOption(pckeVerifier),
		oidc.Nonce(nonce),
		oauth2.SetAuthURLParam("prompt", "select_account"), // Always show account
	)

	// The browser redirects to Google's login page
	redirect(c, http.StatusFound, loginURL)
}

type googleClaims struct {
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

func (a *AuthController) GoogleCallback(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Error from callback URL if any
	if oauthErr := c.Query("error"); oauthErr != "" {
		redirectError(c, errors.New(oauthErr))
		return
	}

	// Verify that the callback is tied to this login attempt
	state := a.Manager.PopString(ctx, "state")
	callbackState := c.Query("state")
	if !equal(state, callbackState) {
		redirectError(c, fmt.Errorf("invalid state. state=%s & callbackState=%s", state, callbackState))
		return
	}

	// Exchange code for id token
	pkceVerifier := a.Manager.PopString(ctx, "pkce_verifier")
	if pkceVerifier == "" {
		redirectError(c, errors.New("missing pcke"))
		return
	}
	code := c.Query("code")
	if code == "" {
		redirectError(c, errors.New("missing authz code"))
		return
	}
	idToken, err := a.exchangeCodeForIdToken(ctx, code, pkceVerifier)
	if err != nil {
		redirectError(c, fmt.Errorf("unable to exchange code for token: %w", err))
		return
	}

	// Validate the id token
	nonce := a.Manager.PopString(ctx, "nonce")
	if !equal(nonce, idToken.Nonce) {
		redirectError(c, fmt.Errorf("invalid nonce. nonce=%s & idtoken nonce=%s", nonce, idToken.Nonce))
		return
	}
	if idToken.Subject == "" {
		redirectError(c, errors.New("id token has no subject"))
		return
	}

	// Unmarshal claims from id token and validate claims
	var claims googleClaims
	if err = idToken.Claims(&claims); err != nil {
		redirectError(c, fmt.Errorf("unable to parse claims: %w", err))
		return
	}
	if !claims.EmailVerified || strings.TrimSpace(claims.Email) == "" {
		redirectError(c, errors.New("email is not verified or non-existent"))
		return
	}

	// Create a user or find it
	userId, err := a.findOrInsertUser(ctx, *idToken, claims)
	if err != nil {
		redirectError(c, fmt.Errorf("unable to find or create user: %w", err))
		return
	}

	// Renew the token and data in session and cookie
	if err = a.Manager.RenewToken(ctx); err != nil {
		redirectError(c, err)
		return
	}
	a.Manager.Put(ctx, "user_id", userId.String())

	redirect(c, http.StatusSeeOther, "http://localhost:3000/dashboard")
}

// exchangeCodeForIdToken exchanges authorization code for the Id token
func (a *AuthController) exchangeCodeForIdToken(
	ctx context.Context,
	authzCode string,
	pkceVerifier string,
) (*oidc.IDToken, error) {
	oauth2Token, err := a.GoogleClient.OAuthConfig.Exchange(
		ctx,
		authzCode,
		oauth2.VerifierOption(pkceVerifier),
	)
	if err != nil {
		return nil, err
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("unable to access id token from oauth2 token")
	}

	idToken, err := a.GoogleClient.Verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, err
	}

	return idToken, nil
}

// findOrInsertUser queries the user from database,
// if it doesn't exist then it will create the user in the database on signup
func (a *AuthController) findOrInsertUser(
	ctx context.Context,
	idToken oidc.IDToken,
	claims googleClaims,
) (uuid.UUID, error) {
	var userId uuid.UUID
	if err := a.DB.QueryRow(
		ctx,
		"select user_id from user_identities where issuer = $1 and subject = $2;",
		idToken.Issuer, idToken.Subject,
	).Scan(&userId); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			insertedUserIdRow := a.DB.QueryRow(
				ctx,
				`with inserted_user as (
					insert into users (email, display_name, avatar_url, role)
					values ($1, $2, $3, 'user')
					returning id
				)
				insert into user_identities (
					user_id, provider, issuer, subject, email, email_verified
				)
				select id, $4, $5, $6, $1, $7 from inserted_user
				returning user_id;`,
				claims.Email,
				claims.Name,
				claims.Picture,
				googleProvider,
				idToken.Issuer,
				idToken.Subject,
				claims.EmailVerified,
			)
			if err = insertedUserIdRow.Scan(&userId); err != nil {
				log.Printf("Insert failed: %v", err)
				return uuid.UUID{}, err
			}
		} else {
			return uuid.UUID{}, err
		}
	}

	return userId, nil
}

func (a *AuthController) GoogleLogout(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Remove the session in the database
	if err := a.Manager.Destroy(ctx); err != nil {
		redirectError(c, err)
		return
	}
	a.Manager.PopString(ctx, "user_id")

	redirect(c, http.StatusSeeOther, "http://localhost:3000")
}
