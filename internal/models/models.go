package models

import (
	"time"

	"github.com/google/uuid"
)

type Identity struct {
	Provider      string
	Issuer        string
	Subject       string
	Name          string
	Picture       string
	Email         string
	EmailVerified bool
}

type User struct {
	ID          uuid.UUID `db:"id" json:"id" format:"uuid"`
	Email       string    `db:"email" json:"email" format:"email"`
	DisplayName *string   `db:"display_name" json:"display_name,omitempty"`
	AvatarURL   *string   `db:"avatar_url" json:"avatar_url,omitempty" format:"uri"`
	Role        string    `db:"role" json:"role" enums:"user,admin"`
	CreatedAt   time.Time `db:"created_at" json:"created_at" format:"date-time"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at" format:"date-time"`
}

type APIKey struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	UserID       uuid.UUID  `db:"user_id" json:"user_id"`
	Project      string     `db:"project" json:"project"`
	Description  string     `db:"description" json:"description"`
	HashedKey    string     `db:"hashed_key" json:"-"`
	EncryptedKey *string    `db:"encrypted_key" json:"-"`
	Internal     bool       `db:"internal" json:"internal"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	RevokedAt    *time.Time `db:"revoked_at" json:"revoked_at,omitempty"`
}
