package dto

import (
	"time"

	"github.com/google/uuid"

	"loviary.app/backend/internal/domain/shared"
	"loviary.app/backend/internal/domain/users"
)

// UserResponse is the unified user model returned by ALL endpoints that include user data.
// This includes: POST /auth/login, POST /auth/refresh, GET /users/me, PATCH /users/me.
// HasCouple indicates whether the user currently has an active couple relationship,
// which allows the mobile client to navigate without an additional API call.
type UserResponse struct {
	ID            uuid.UUID      `json:"id"`
	Username      string         `json:"username"`
	Email         string         `json:"email"`
	FirstName     *string        `json:"first_name,omitempty"`
	LastName      *string        `json:"last_name,omitempty"`
	DateOfBirth   *time.Time     `json:"date_of_birth,omitempty"`
	Gender        *shared.Gender `json:"gender,omitempty"`
	Language      string         `json:"language"`
	KeyCouple     *string        `json:"key_couple,omitempty"`
	AvatarURL     *string        `json:"avatar_url,omitempty"`
	IsActive      bool           `json:"is_active"`
	EmailVerified bool           `json:"email_verified"`
	HasCouple     bool           `json:"has_couple"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// UserToResponse converts a User domain model to the unified UserResponse DTO.
// hasCouple must be resolved by the caller (e.g. by querying the couple repository).
func UserToResponse(user *users.User, hasCouple bool) UserResponse {
	return UserResponse{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		DateOfBirth:   user.DateOfBirth,
		Gender:        user.Gender,
		Language:      user.Language,
		KeyCouple:     user.KeyCouple,
		AvatarURL:     user.AvatarURL,
		IsActive:      user.IsActive,
		EmailVerified: user.EmailVerified,
		HasCouple:     hasCouple,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}
