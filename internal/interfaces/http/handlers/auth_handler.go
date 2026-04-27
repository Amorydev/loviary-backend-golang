package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	appAuth "loviary.app/backend/internal/application/auth"
	"loviary.app/backend/internal/application/verification"
	"loviary.app/backend/internal/domain/users"
	"loviary.app/backend/internal/interfaces/http/middleware"
	apperrors "loviary.app/backend/pkg/errors"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService        *appAuth.Service
	verificationService *verification.Service
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *appAuth.Service, verificationService *verification.Service) *AuthHandler {
	return &AuthHandler{
		authService:        authService,
		verificationService: verificationService,
	}
}

// RegisterRequest represents registration request
type RegisterRequest struct {
	Email     string  `json:"email" binding:"required,email"`
	Username  string  `json:"username" binding:"required,min=3,max=50,alphanum"`
	Password  string  `json:"password" binding:"required,min=8,max=100"`
	Language  string  `json:"language"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
}

// Validate validates the register request
func (r *RegisterRequest) Validate() error {
	if r.Language == "" {
		r.Language = "en"
	}
	return nil
}

// VerifyEmailRequest represents email verification request
type VerifyEmailRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	Code   string    `json:"code" binding:"required,len=6"`
}

// Validate validates the verify email request
func (r *VerifyEmailRequest) Validate() error {
	return nil
}

// ResendVerificationRequest represents resend verification request
type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// Validate validates the resend verification request
func (r *ResendVerificationRequest) Validate() error {
	return nil
}

// LoginRequest represents login request
type LoginRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	DeviceInfo string `json:"device_info"`
}

// Validate validates the login request
func (r *LoginRequest) Validate() error {
	return nil
}

// RefreshRequest represents refresh token request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Validate validates the refresh request
func (r *RefreshRequest) Validate() error {
	return nil
}

// LogoutRequest represents logout request
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Register handles user registration
// @Summary Register a new user
// @Description Create a new user account with email, username and password
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   request  body  handlers.RegisterRequest  true  "Registration request"
// @Success  201  {object}  handlers.RegisterResponse "Successfully registered"
// @Failure  400  {object}  handlers.ErrorResponse "Invalid input"
// @Failure  409  {object}  handlers.ErrorResponse "Email or username already exists"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", err.Error()))
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, apperrors.New("VALIDATION_ERROR", err.Error()))
		return
	}

	user, err := h.authService.Register(c.Request.Context(), appAuth.RegisterRequest{
		Email:     req.Email,
		Username:  req.Username,
		Password:  req.Password,
		Language:  req.Language,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		switch {
		case errors.Is(err, users.ErrDuplicateEmail):
			c.JSON(http.StatusConflict, apperrors.New("EMAIL_EXISTS", "Email already registered"))
		case errors.Is(err, users.ErrDuplicateUsername):
			c.JSON(http.StatusConflict, apperrors.New("USERNAME_EXISTS", "Username already taken"))
		default:
			c.JSON(http.StatusInternalServerError, apperrors.New("INTERNAL_ERROR", "Failed to register user"))
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"user_id":    user.ID,
			"email":      user.Email,
			"key_couple": user.KeyCouple,
		},
		"message": "Vui lòng kiểm tra email để xác nhận tài khoản.",
	})
}

// Login handles user login
// @Summary Login user
// @Description Authenticate user with email and password, return access and refresh tokens
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   request  body  handlers.LoginRequest  true  "Login credentials"
// @Success  200  {object}  handlers.LoginResponse "Login successful with tokens"
// @Failure  400  {object}  handlers.ErrorResponse "Invalid input"
// @Failure  401  {object}  handlers.ErrorResponse "Invalid credentials or email not verified"
// @Failure  403  {object}  handlers.ErrorResponse "Account disabled"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", err.Error()))
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, apperrors.New("VALIDATION_ERROR", err.Error()))
		return
	}

	result, err := h.authService.Login(c.Request.Context(), appAuth.LoginInput{
		Email:      req.Email,
		Password:   req.Password,
		DeviceInfo: req.DeviceInfo,
	})
	if err != nil {
		switch {
		case errors.Is(err, appAuth.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, apperrors.New("INVALID_CREDENTIALS", "Invalid email or password"))
		case errors.Is(err, appAuth.ErrEmailNotVerified):
			c.JSON(http.StatusForbidden, apperrors.New("EMAIL_NOT_VERIFIED", "Please verify your email first"))
		case errors.Is(err, appAuth.ErrAccountDisabled):
			c.JSON(http.StatusForbidden, apperrors.New("ACCOUNT_DISABLED", "Account is disabled"))
		default:
			c.JSON(http.StatusInternalServerError, apperrors.New("INTERNAL_ERROR", "Failed to login"))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"access_token":  result.AccessToken,
			"refresh_token": result.RefreshToken,
			"expires_in":    result.ExpiresIn,
			"user": gin.H{
				"user_id":        result.User.ID,
				"username":       result.User.Username,
				"email":          result.User.Email,
				"first_name":     result.User.FirstName,
				"last_name":      result.User.LastName,
				"avatar_url":     result.User.AvatarURL,
				"is_active":      result.User.IsActive,
				"email_verified": result.User.EmailVerified,
			},
		},
	})
}

// Refresh handles token refresh
// @Summary Refresh access token
// @Description Get new access and refresh tokens using a valid refresh token
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   request  body  handlers.RefreshRequest  true  "Refresh token"
// @Success  200  {object}  handlers.RefreshResponse "New tokens issued"
// @Failure  400  {object}  handlers.ErrorResponse "Invalid input"
// @Failure  401  {object}  handlers.ErrorResponse "Token expired, revoked or invalid"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", err.Error()))
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, apperrors.New("VALIDATION_ERROR", err.Error()))
		return
	}

	result, err := h.authService.Refresh(c.Request.Context(), appAuth.RefreshTokenInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		switch {
		case errors.Is(err, appAuth.ErrTokenExpired):
			c.JSON(http.StatusUnauthorized, apperrors.New("TOKEN_EXPIRED", "Refresh token has expired"))
		case errors.Is(err, appAuth.ErrTokenRevoked):
			c.JSON(http.StatusUnauthorized, apperrors.New("TOKEN_REVOKED", "Refresh token has been revoked"))
		case errors.Is(err, appAuth.ErrInvalidToken):
			c.JSON(http.StatusUnauthorized, apperrors.New("TOKEN_INVALID", "Invalid refresh token"))
		default:
			c.JSON(http.StatusInternalServerError, apperrors.New("INTERNAL_ERROR", "Failed to refresh token"))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"access_token":  result.AccessToken,
			"refresh_token": result.RefreshToken,
			"expires_in":    result.ExpiresIn,
		},
	})
}

// Logout handles user logout
// @Summary Logout user
// @Description Invalidate refresh token and logout from current session
// @Tags auth
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param   request  body  handlers.LogoutRequest  true  "Logout request with refresh token"
// @Success  200  {object}  handlers.SuccessResponse "Logged out successfully"
// @Failure  400  {object}  handlers.ErrorResponse "Invalid input"
// @Failure  401  {object}  handlers.ErrorResponse "Not authenticated"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", err.Error()))
		return
	}

	// Verify user is authenticated
	if _, exists := middleware.GetUserID(c); !exists {
		c.JSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
		return
	}

	if err := h.authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, apperrors.New("INTERNAL_ERROR", "Failed to logout"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// LogoutAll handles logout from all devices
// @Summary Logout all sessions
// @Description Invalidate all refresh tokens for the current user
// @Tags auth
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success  200  {object}  handlers.SuccessResponse "Logged out from all devices"
// @Failure  401  {object}  handlers.ErrorResponse "Not authenticated"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
		return
	}

	if err := h.authService.LogoutAll(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, apperrors.New("INTERNAL_ERROR", "Failed to logout all sessions"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// VerifyEmail verifies a user's email with the 6-digit code
// @Summary Verify email
// @Description Verify user's email with 6-digit verification code
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   request  body  handlers.VerifyEmailRequest  true  "Verification code"
// @Success  200  {object}  handlers.SuccessResponse "Email verified successfully"
// @Failure  400  {object}  handlers.ErrorResponse "Invalid input"
// @Failure  404  {object}  handlers.ErrorResponse "Verification not found"
// @Failure  410  {object}  handlers.ErrorResponse "Code expired"
// @Failure  409  {object}  handlers.ErrorResponse "Email already verified"
// @Router   /auth/verify-email [post]
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", err.Error()))
		return
	}

	// Gọi verification service để xác thực mã
	_, err := h.verificationService.VerifyCode(c.Request.Context(), req.Code)
	if err != nil {
		switch {
		case errors.Is(err, verification.ErrNotFound):
			c.JSON(http.StatusNotFound, apperrors.New("VERIFICATION_NOT_FOUND", err.Error()))
		case errors.Is(err, verification.ErrExpired):
			c.JSON(http.StatusGone, apperrors.New("CODE_EXPIRED", err.Error()))
		case errors.Is(err, verification.ErrAlreadyVerified):
			c.JSON(http.StatusConflict, apperrors.New("ALREADY_VERIFIED", err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, apperrors.New("INTERNAL_ERROR", "Failed to verify email"))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Email verified successfully",
	})
}

// ResendVerification resends the verification email
// @Summary Resend verification email
// @Description Resend verification email with new code
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   request  body  handlers.ResendVerificationRequest  true  "Email address"
// @Success  200  {object}  handlers.SuccessResponse "Verification email sent"
// @Failure  400  {object}  handlers.ErrorResponse "Invalid input"
// @Failure  404  {object}  handlers.ErrorResponse "User not found or already verified"
// @Failure  409  {object}  handlers.ErrorResponse "Verification pending or resend too fast"
// @Router   /auth/resend-verification [post]
func (h *AuthHandler) ResendVerification(c *gin.Context) {
	var req ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", err.Error()))
		return
	}

	// Kiểm tra user tồn tại
	user, err := h.authService.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			c.JSON(http.StatusNotFound, apperrors.New("USER_NOT_FOUND", "User not found"))
		} else {
			c.JSON(http.StatusInternalServerError, apperrors.New("INTERNAL_ERROR", "Failed to get user"))
		}
		return
	}

	if user.EmailVerified {
		c.JSON(http.StatusConflict, apperrors.New("ALREADY_VERIFIED", "Email already verified"))
		return
	}

	// Kiểm tra rate limit
	canResend, next, err := h.verificationService.CanResend(c.Request.Context(), user.ID)
	if err != nil {
		// Nếu có lỗi khác (như không tìm thấy verification), cho phép tạo mới
		if errors.Is(err, verification.ErrNotFound) || errors.Is(err, verification.ErrExpired) {
			// Tiếp tục để tạo verification mới
			canResend = true
		} else if errors.Is(err, verification.ErrAlreadyVerified) {
			c.JSON(http.StatusConflict, apperrors.New("ALREADY_VERIFIED", "Email already verified"))
			return
		} else {
			c.JSON(http.StatusConflict, apperrors.New("RESEND_ERROR", err.Error()))
			return
		}
	}

	if !canResend {
		waitTime := time.Until(next).Round(time.Second)
		c.JSON(http.StatusTooManyRequests, apperrors.New("RATE_LIMIT_EXCEEDED",
			fmt.Sprintf("Please wait %v before resending", waitTime)))
		return
	}

	// Gửi lại verification email
	if err := h.verificationService.CreateVerification(c.Request.Context(), user.ID, user.Email); err != nil {
		c.JSON(http.StatusInternalServerError, apperrors.New("INTERNAL_ERROR", "Failed to resend verification"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Verification email sent",
	})
}
