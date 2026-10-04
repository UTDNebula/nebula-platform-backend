package config

import (
	"context"
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

type OIDCSettings struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	IssuerURL    string
	Scopes       []string
}

// Initialize the Google's OIDC s
func NewGoogleSettings() (*OIDCSettings, error) {
	s := &OIDCSettings{
		IssuerURL: "https://accounts.google.com",
		Scopes:    []string{"openid", "email", "profile"},
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
	if s.ClientSecret == "" {
		return nil, fmt.Errorf("Missing GOOGLE_REDIRECT_URL")
	}

	return s, nil
}

// Initialize the Gihub's OIDC s
func NewGithubSettings() (*OIDCSettings, error) {
	return nil, nil
}

type OIDCClient struct {
	OAuthConfig *oauth2.Config

	// The verifier ensures that an ID token was issued specifically for our application.
	Verifier *oidc.IDTokenVerifier
}

// Initialize the OIDC client from the given s
func NewOIDCClient(s OIDCSettings) (*OIDCClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Contact the issuer's discovery endpoint.
	// This discovers its authorization URL, token URL, and signing keys.
	provider, err := oidc.NewProvider(ctx, s.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("discover OIDC provider: %w", err)
	}

	oauthConfig := &oauth2.Config{
		ClientID:     s.ClientID,
		ClientSecret: s.ClientSecret,
		RedirectURL:  s.RedirectURL,

		// These endpoints were discovered from the issuer.
		Endpoint: provider.Endpoint(),

		Scopes: s.Scopes,
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: s.ClientID,
	})

	return &OIDCClient{
		OAuthConfig: oauthConfig,
		Verifier:    verifier,
	}, nil
}

// Initialize the session manager for the token
func NewSessionManager(database *pgxpool.Pool) (*scs.SessionManager, error) {
	manager := scs.New()
	// Session manager will be connected to the database
	manager.Store = pgxstore.New(database)

	manager.Lifetime = 1 * time.Hour
	manager.IdleTimeout = 15 * time.Minute
	manager.HashTokenInStore = true

	manager.Cookie.Name = "nebula_session"
	manager.Cookie.Path = "/"
	manager.Cookie.HttpOnly = true
	manager.Cookie.Persist = true
	manager.Cookie.SameSite = http.SameSiteLaxMode
	manager.Cookie.Secure = false

	return manager, nil
}
