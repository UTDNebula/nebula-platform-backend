package controllers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/UTDNebula/nebula-platform-backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserController struct {
	db *pgxpool.Pool
}

func NewUserController(db *pgxpool.Pool) *UserController {
	return &UserController{
		db: db,
	}
}

func (u *UserController) Profile(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	userID := c.GetString("user_id")
	var user models.User

	userRow := u.db.QueryRow(ctx,
		`select id, email, display_name, avatar_url, created_at, updated_at
		from users where id = $1;`,
		userID,
	)
	if err := userRow.Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respond(c, http.StatusNotFound, "error", errors.New("user not found"))
		} else {
			respondInternalError(c, err)
		}
		return
	}

	respond(c, http.StatusAccepted, "data", user)
}
