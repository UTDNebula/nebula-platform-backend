package config

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
)

type AuthSettings struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

type OIDCSettings struct {
	AuthSettings
	IssuerURL string
}

// Initialize the Google's OIDC settings
func NewGoogleSettings() (*OIDCSettings, error) {
	s := &OIDCSettings{
		AuthSettings: AuthSettings{
			Scopes: []string{"openid", "email", "profile"},
		},
		IssuerURL: "https://accounts.google.com",
	}

	s.ClientID = strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID"))
	if s.ClientID == "" {
		return nil, fmt.Errorf("Missing GOOGLE_CLIENT_ID")
	}

	s.ClientSecret = strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET"))
	if s.ClientSecret == "" {
		return nil, fmt.Errorf("Missing GOOGLE_CLIENT_SECRET")
	}

	s.RedirectURL = strings.TrimSpace(os.Getenv("GOOGLE_REDIRECT_URL"))
	if s.RedirectURL == "" {
		return nil, fmt.Errorf("Missing GOOGLE_REDIRECT_URL")
	}

	return s, nil
}

// TODO: Implement the new settings for other provider

type Identity struct {
	Provider      string
	Issuer        string
	Subject       string
	Name          string
	Picture       string
	Email         string
	EmailVerified bool
}

// IdentityFetcher uses an access token to fetch user information without an ID token or nonce.
type IdentityFetcher func(ctx context.Context, token *oauth2.Token) (Identity, error)

// TODO: Implement any non-OIDC identify fetcher

type OIDCIdentityFetcher func(ctx context.Context, token *oauth2.Token, nonce string) (Identity, error)

// Reusable function to fetch identity from OIDC auth provider
func fetchOIDCIdentity(provider string, verifier *oidc.IDTokenVerifier) OIDCIdentityFetcher {
	return func(ctx context.Context, token *oauth2.Token, nonce string) (Identity, error) {
		// Fetch id token from oauth2 token
		rawIDToken, ok := token.Extra("id_token").(string)
		if !ok {
			return Identity{}, fmt.Errorf("unable to access id token from oauth2 token")
		}
		idToken, err := verifier.Verify(ctx, rawIDToken)
		if err != nil {
			return Identity{}, err
		}

		// Validate the id token
		if !TokenEqual(nonce, idToken.Nonce) {
			return Identity{}, errors.New("invalid nonce")
		}
		if idToken.Subject == "" {
			return Identity{}, errors.New("id token has no subject")
		}

		// Unmarshal claims from id token
		var claims struct {
			Name          string `json:"name"`
			Picture       string `json:"picture"`
			Email         string `json:"email"`
			EmailVerified bool   `json:"email_verified"`
		}
		if err = idToken.Claims(&claims); err != nil {
			return Identity{}, fmt.Errorf("unable to parse claims: %w", err)
		}

		// Validate claims
		if !claims.EmailVerified || strings.TrimSpace(claims.Email) == "" {
			return Identity{}, errors.New("email is not verified or non-existent")
		}

		return Identity{
			Provider:      provider,
			Issuer:        idToken.Issuer,
			Subject:       idToken.Subject,
			Name:          claims.Name,
			Picture:       claims.Picture,
			Email:         claims.Email,
			EmailVerified: claims.EmailVerified,
		}, nil
	}
}

type AuthProvider struct {
	// Client Id, client secret, redirect URL, etc.
	OAuthConfig *oauth2.Config

	// For example: Always showing the select account screen, etc.
	AuthOptions []oauth2.AuthCodeOption

	// Does this provider supports OpenID Connect?
	SupportOIDC bool

	// Does this provider support PKCE verifier when exchanging authz code for oauth2 token?
	SupportPKCE bool

	// Method to get identify from oauth2 token if OIDC is not supported, null otherwise
	FetchIdentity IdentityFetcher

	// Method to get identify from oauth2 token if OIDC is supported, null otherwise
	FetchOIDCIdentity OIDCIdentityFetcher
}

// TODO: Implement NewAuthProvider constructs a non-OIDC provider separately or
// constructs both non-OIDC provider and OIDC provider, technically replacing NewOIDCAuthProvider as well

// NewOIDCAuthProvider constructs an OIDC provider
func NewOIDCAuthProvider(
	settings OIDCSettings,
	options []oauth2.AuthCodeOption,
	supportPKCE bool,
	name string,
) (*AuthProvider, error) {
	if strings.TrimSpace(settings.IssuerURL) == "" {
		return nil, errors.New("OIDC provider requires an issuer URL")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	issuer, err := oidc.NewProvider(ctx, settings.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("discover OIDC provider: %w", err)
	}

	verifier := issuer.Verifier(&oidc.Config{
		ClientID: settings.ClientID,
	})

	provider := &AuthProvider{
		OAuthConfig: &oauth2.Config{
			ClientID:     settings.ClientID,
			ClientSecret: settings.ClientSecret,
			RedirectURL:  settings.RedirectURL,
			Scopes:       settings.Scopes,
			Endpoint:     issuer.Endpoint(),
		},
		AuthOptions:       options,
		SupportOIDC:       true,
		SupportPKCE:       supportPKCE,
		FetchOIDCIdentity: fetchOIDCIdentity(name, verifier),
	}

	return provider, nil
}

// Initialize the session manager for the token
func NewSessionManager(database *pgxpool.Pool) (*scs.SessionManager, error) {
	manager := scs.New()
	// Session manager will be connected to the database
	manager.Store = pgxstore.New(database)

	manager.Lifetime = 1 * time.Hour
	manager.IdleTimeout = 30 * time.Minute
	manager.HashTokenInStore = true

	manager.Cookie.Name = "nebula_session"
	manager.Cookie.Path = "/"
	manager.Cookie.HttpOnly = true
	manager.Cookie.Persist = true
	manager.Cookie.SameSite = http.SameSiteLaxMode
	manager.Cookie.Secure = false

	return manager, nil
}

func TokenEqual(a string, b string) bool {
	return a != "" && b != "" && subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
