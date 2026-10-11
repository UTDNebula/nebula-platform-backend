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
	DB *pgxpool.Pool
}

func NewUserController(db *pgxpool.Pool) *UserController {
	return &UserController{DB: db}
}

func (u *UserController) Profile(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	userID := c.GetString("user_id")
	var user models.User

	userRow := u.DB.QueryRow(ctx,
		`select id, email, display_name, avatar_url, role, created_at, updated_at
		from users where id = $1;`,
		userID,
	)
	if err := userRow.Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.AvatarURL,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			Respond(c, http.StatusNotFound, "error", errors.New("user not found"))
		} else {
			RespondInternalError(c, err)
		}
		return
	}

	Respond(c, http.StatusAccepted, "data", user)
}

func (u *UserController) ListUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	rows, err := u.DB.Query(ctx,
		`select id, email, display_name, avatar_url, role, created_at, updated_at
		from users
		order by created_at desc;`,
	)
	if err != nil {
		RespondInternalError(c, err)
		return
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.DisplayName,
			&user.AvatarURL,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			RespondInternalError(c, err)
			return
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		RespondInternalError(c, err)
		return
	}

	Respond(c, http.StatusOK, "data", users)
}
