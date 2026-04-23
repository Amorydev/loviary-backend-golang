package dto

import (
    "time"

    "github.com/google/uuid"

    "loviary.app/backend/internal/domain/users"
    "loviary.app/backend/internal/domain/shared"
)

// UserResponse represents user response DTO
type UserResponse struct {
    ID            uuid.UUID  `json:"id"`
    Username      string     `json:"username"`
    Email         string     `json:"email"`
    FirstName     *string    `json:"first_name"`
    LastName      *string    `json:"last_name"`
    DateOfBirth   *time.Time `json:"date_of_birth"`
    Gender        *shared.Gender `json:"gender"`
    Language      string     `json:"language"`
    KeyCouple     *string    `json:"key_couple,omitempty"`
    AvatarURL     *string    `json:"avatar_url"`
    IsActive      bool       `json:"is_active"`
    EmailVerified bool       `json:"email_verified"`
    CreatedAt     time.Time  `json:"created_at"`
    UpdatedAt     time.Time  `json:"updated_at"`
}

// UserToResponse converts a User to response DTO
func UserToResponse(user *users.User) UserResponse {
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
        CreatedAt:     user.CreatedAt,
        UpdatedAt:     user.UpdatedAt,
    }
}
