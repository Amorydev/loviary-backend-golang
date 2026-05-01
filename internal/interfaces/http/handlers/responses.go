package handlers

import (
	"github.com/google/uuid"

	"loviary.app/backend/internal/interfaces/http/dto"
)

// Ensure uuid is used (RegisterData uses uuid.UUID).
var _ = uuid.UUID{}

// SuccessResponse is a generic success response with no data payload.
type SuccessResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message,omitempty" example:"Operation completed successfully"`
}

// ----- Auth Responses -----

// RegisterData holds info returned after registration.
type RegisterData struct {
	UserID    uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string    `json:"email" example:"user@loviary.app"`
	KeyCouple *string   `json:"key_couple,omitempty" example:"ABC123"`
}

// RegisterResponse is the response body for POST /auth/register.
type RegisterResponse struct {
	Success bool         `json:"success" example:"true"`
	Data    RegisterData `json:"data"`
	Message string       `json:"message" example:"Vui lòng kiểm tra email để xác nhận tài khoản."`
}

// LoginData holds tokens and user data returned after login.
type LoginData struct {
	AccessToken  string           `json:"access_token"`
	RefreshToken string           `json:"refresh_token"`
	ExpiresIn    int64            `json:"expires_in" example:"900"` // seconds
	HasCouple    bool             `json:"has_couple"`
	User         dto.UserResponse `json:"user"`
}

// LoginResponse is the response body for POST /auth/login.
type LoginResponse struct {
	Success bool      `json:"success" example:"true"`
	Data    LoginData `json:"data"`
}

// TokenData holds the tokens returned by refresh.
type TokenData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in" example:"900"` // seconds
}

// RefreshResponse is the response body for POST /auth/refresh.
type RefreshResponse struct {
	Success bool      `json:"success" example:"true"`
	Data    TokenData `json:"data"`
}

// ----- User Responses -----

// UserResponse re-exports dto.UserResponse for swagger visibility from the handlers package.
type UserResponse = dto.UserResponse

// ----- List Responses -----

// MoodListResponse is the paginated list response for GET /moods/history.
type MoodListResponse struct {
	Moods []dto.MoodResponse `json:"moods"`
	Count int                `json:"count" example:"5"`
}

// MemoryListResponse is the paginated list response for GET /memories.
type MemoryListResponse struct {
	Memories []dto.MemoryResponse `json:"memories"`
	Count    int                  `json:"count" example:"10"`
	Limit    int                  `json:"limit" example:"20"`
	Offset   int                  `json:"offset" example:"0"`
}

// ReminderListResponse is the list response for GET /reminders.
type ReminderListResponse struct {
	Reminders []dto.ReminderResponse `json:"reminders"`
	Count     int                    `json:"count" example:"3"`
}

// StreakListResponse is the list response for GET /streaks/me.
type StreakListResponse struct {
	Streaks []dto.StreakResponse `json:"streaks"`
	Count   int                  `json:"count" example:"2"`
}

// ----- Couple Responses -----

// CoupleResponse represents the couple domain object returned by GET /couples/me.
type CoupleResponse struct {
	CoupleID              string  `json:"couple_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	User1ID               string  `json:"user1_id"`
	User2ID               *string `json:"user2_id,omitempty"`
	CoupleName            *string `json:"couple_name,omitempty"`
	RelationshipStartDate *string `json:"relationship_start_date,omitempty"`
	Status                string  `json:"status" example:"active"`
	RelationshipType      string  `json:"relationship_type" example:"dating"`
	InvitationExpiresAt   *string `json:"invitation_expires_at,omitempty"`
	CreatedAt             string  `json:"created_at"`
	UpdatedAt             string  `json:"updated_at"`
}

// MessageResponse is a simple message-only success response.
type MessageResponse struct {
	Message string `json:"message" example:"Operation completed"`
}
