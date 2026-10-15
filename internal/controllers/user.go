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

// Profile returns the currently authenticated user's profile.
// @Summary Get current user profile
// @Description Requires an authenticated browser session. Returns the current user with HTTP 202, matching the handler's existing behavior.
// @Tags Users
// @Produce json
// @Security SessionAuth
// @Success 202 {object} APIResponse[models.User] "Current user profile"
// @Failure 401 {object} APIResponse[string] "Missing authenticated session"
// @Failure 404 {object} APIResponse[any] "User not found; Data is an empty object in the current implementation"
// @Failure 500 {object} APIResponse[string] "Database or other internal error"
// @Router /api/users/profile [get]
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

// ListUsers returns all users for an authenticated administrator.
// @Summary List users
// @Description Requires an authenticated session with the admin role. Returns all users ordered by creation time, newest first. The current middleware returns HTTP 401 for both missing sessions and non-admin users.
// @Tags Users
// @Produce json
// @Security SessionAuth
// @Success 200 {object} APIResponse[[]models.User] "Users ordered by creation time descending"
// @Failure 401 {object} APIResponse[string] "Missing authenticated session or admin role"
// @Failure 500 {object} APIResponse[string] "Database or other internal error"
// @Router /api/admin/users/ [get]
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
