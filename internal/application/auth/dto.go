package auth

import (
    "github.com/go-playground/validator/v10"
)

var validate = validator.New()

// RegisterRequest cho đăng ký tài khoản
type RegisterRequest struct {
    Username  string  `json:"username" validate:"required,min=3,max=50,alphanum"`
    Email     string  `json:"email" validate:"required,email"`
    Password  string  `json:"password" validate:"required,min=8,max=100"`
    FirstName *string `json:"first_name,omitempty" validate:"omitempty,max=50"`
    LastName  *string `json:"last_name,omitempty" validate:"omitempty,max=50"`
    Language  string  `json:"language" validate:"required,oneof=vi en"`
}

// Validate kiểm tra input
func (r *RegisterRequest) Validate() error {
    return validate.Struct(r)
}

// LoginRequest cho đăng nhập
type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
    FCMToken *string `json:"fcm_token,omitempty"`
    Platform *string `json:"platform,omitempty" validate:"omitempty,oneof=ios android web"`
    DeviceName *string `json:"device_name,omitempty"`
}

func (r *LoginRequest) Validate() error {
    return validate.Struct(r)
}

// RefreshRequest cho lấy token mới
type RefreshRequest struct {
    RefreshToken string `json:"refresh_token" validate:"required"`
}

func (r *RefreshRequest) Validate() error {
    return validate.Struct(r)
}

// LogoutRequest cho đăng xuất
type LogoutRequest struct {
    RefreshToken string `json:"refresh_token" validate:"required"`
}

func (r *LogoutRequest) Validate() error {
    return validate.Struct(r)
}

// TokenResponse chứa access và refresh token
type TokenResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int    `json:"expires_in"` // seconds
}
