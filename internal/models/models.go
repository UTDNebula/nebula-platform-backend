package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	Id        bson.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Email     string        `json:"email" bson:"email" validate:"required,email"`
	CreatedAt time.Time     `json:"created_at" bson:"created_at"`
}

type APIKey struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	UserID       bson.ObjectID `bson:"user_id" json:"user_id"`
	Project      string        `bson:"project" json:"project"`
	Description  string        `bson:"description" json:"description"`
	HashedKey    string        `bson:"hashed_key" json:"-"`
	EncryptedKey string        `bson:"encrypted_key,omitempty" json:"-"`
	Internal     bool          `bson:"internal" json:"internal"`
	Active       bool          `bson:"active" json:"active"`
	CreatedAt    time.Time     `bson:"created_at" json:"created_at"`
	RevokedAt    *time.Time    `bson:"revoked_at,omitempty" json:"revoked_at,omitempty"`
}
