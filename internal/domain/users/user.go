package users

import (
    "errors"
    "time"

    "github.com/google/uuid"

    "loviary.app/backend/internal/domain/shared"
)

// Error constants
var (
    ErrNotFound          = errors.New("NOT_FOUND")
    ErrDuplicateEmail    = errors.New("DUPLICATE_EMAIL")
    ErrDuplicateUsername = errors.New("DUPLICATE_USERNAME")
    ErrInvalidPassword   = errors.New("INVALID_PASSWORD")
)

// User represents a user in the system
type User struct {
    ID            uuid.UUID  `db:"user_id" json:"user_id"`
    Username      string     `db:"username" json:"username" validate:"required,min=3,max=50,alphanum"`
    Email         string     `db:"email" json:"email" validate:"required,email,max=100"`
    PasswordHash  string     `db:"password_hash" json:"-"`
    FirstName     *string    `db:"first_name" json:"first_name" validate:"max=50"`
    LastName      *string    `db:"last_name" json:"last_name" validate:"max=50"`
    DateOfBirth   *time.Time `db:"date_of_birth" json:"date_of_birth"`
    Gender        *shared.Gender `db:"gender" json:"gender" validate:"omitempty,oneof=male female other prefer_not"`
    Language      string     `db:"language" json:"language" validate:"max=10"`
    KeyCouple     *string    `db:"key_couple" json:"key_couple,omitempty"`
    AvatarURL     *string    `db:"avatar_url" json:"avatar_url"`
    IsActive      bool       `db:"is_active" json:"is_active"`
    EmailVerified bool       `db:"email_verified" json:"email_verified"`
    CreatedAt     time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

// IsProfileComplete checks if user has filled basic profile info
func (u *User) IsProfileComplete() bool {
    return u.FirstName != nil && u.LastName != nil && u.DateOfBirth != nil
}

// DisplayName returns the user's full name or username fallback
func (u *User) DisplayName() string {
    if u.FirstName != nil && u.LastName != nil {
        return *u.FirstName + " " + *u.LastName
    }
    if u.FirstName != nil {
        return *u.FirstName
    }
    return u.Username
}
