package controllers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
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

type AuthController struct {
	// List of available auth providers
	AuthProviders map[string]config.AuthProvider

	// Login session manager
	Manager *scs.SessionManager

	DB *pgxpool.Pool
}

// Initialize the auth controllers with OIDC client and session manager
func NewAuthController(
	authProvider map[string]config.AuthProvider,
	sessionManager *scs.SessionManager,
	db *pgxpool.Pool,
) *AuthController {
	return &AuthController{
		AuthProviders: authProvider,
		Manager:       sessionManager,
		DB:            db,
	}
}

func (a *AuthController) Login(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	name := c.Param("provider")
	provider, ok := a.AuthProviders[name]
	if !ok {
		redirectError(c, fmt.Errorf("unsupported auth provider"))
		return
	}
	a.Manager.Put(ctx, "provider", name)

	// state verifies if callback ties to this login
	state, err := generateToken()
	if err != nil {
		redirectError(c, fmt.Errorf("Unable to create state: %w", err))
		return
	}
	a.Manager.Put(ctx, name+"_state", state)

	opts := []oauth2.AuthCodeOption{}

	// PKCE verifier for exchanging code for oauth2 token, if supported by the provider
	if provider.SupportPKCE {
		pckeVerifier, err := generateToken()
		if err != nil {
			redirectError(c, fmt.Errorf("Unable to create pckeVerifier: %w", err))
			return
		}
		a.Manager.Put(ctx, name+"_pkce_verifier", pckeVerifier)
		opts = append(opts, oauth2.S256ChallengeOption(pckeVerifier))
	}

	// nonce verifies ID token, if the provider supports OIDC
	if provider.SupportOIDC {
		nonce, err := generateToken()
		if err != nil {
			redirectError(c, fmt.Errorf("Unable to create nonce: %w", err))
			return
		}
		a.Manager.Put(ctx, name+"_nonce", nonce)
		opts = append(opts, oidc.Nonce(nonce))
	}

	opts = append(opts, provider.AuthOptions...)

	// Build the URL and navigate to consent screen
	loginURL := provider.OAuthConfig.AuthCodeURL(state, opts...)

	redirect(c, http.StatusFound, loginURL)
}

func (a *AuthController) Callback(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Error from callback URL if any
	if oauthErr := c.Query("error"); oauthErr != "" {
		redirectError(c, errors.New(oauthErr))
		return
	}

	name := a.Manager.PopString(ctx, "provider")
	if name == "" || name != c.Param("provider") {
		redirectError(c, errors.New("invalid callback provider"))
		return
	}

	// Verify that the callback is tied to this login attempt
	state := a.Manager.PopString(ctx, name+"_state")
	callbackState := c.Query("state")
	if !config.TokenEqual(state, callbackState) {
		redirectError(c, errors.New("invalid state"))
		return
	}

	provider, ok := a.AuthProviders[name]
	if !ok {
		redirectError(c, errors.New("unsupported auth provider"))
		return
	}

	// Exchange code for oauth2 token
	code := c.Query("code")
	if code == "" {
		redirectError(c, errors.New("missing authz code"))
		return
	}

	opts := []oauth2.AuthCodeOption{}

	if provider.SupportPKCE {
		pkceVerifier := a.Manager.PopString(ctx, name+"_pkce_verifier")
		if pkceVerifier == "" {
			redirectError(c, errors.New("missing PKCE verifier"))
			return
		}

		opts = append(opts, oauth2.VerifierOption(pkceVerifier))
	}

	token, err := provider.OAuthConfig.Exchange(ctx, code, opts...)
	if err != nil {
		redirectError(c, fmt.Errorf("unable to exchange code for token: %w", err))
		return
	}

	// Get the identify from the token
	var identity config.Identity
	if provider.SupportOIDC {
		nonce := a.Manager.PopString(ctx, name+"_nonce")
		identity, err = provider.FetchOIDCIdentity(ctx, token, nonce)
	} else {
		identity, err = provider.FetchIdentity(ctx, token)
	}
	if err != nil {
		redirectError(c, err)
		return
	}

	// Find or create the user based on the identity
	userId, err := a.findOrInsertUser(ctx, identity)
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

// findOrInsertUser queries the user from database,
// if it doesn't exist then it will create the user in the database on sign-up
func (a *AuthController) findOrInsertUser(ctx context.Context, identity config.Identity) (uuid.UUID, error) {
	var userId uuid.UUID
	if err := a.DB.QueryRow(
		ctx,
		"select user_id from user_identities where issuer = $1 and subject = $2;",
		identity.Issuer, identity.Subject,
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
				identity.Email,
				identity.Name,
				identity.Picture,
				identity.Provider,
				identity.Issuer,
				identity.Subject,
				identity.EmailVerified,
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

func (a *AuthController) Logout(c *gin.Context) {
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
