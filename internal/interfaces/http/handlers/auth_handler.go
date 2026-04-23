package handlers

import (
    "errors"
    "net/http"

    "github.com/gin-gonic/gin"

    appAuth "loviary.app/backend/internal/application/auth"
    "loviary.app/backend/internal/domain/users"
    "loviary.app/backend/internal/interfaces/http/middleware"
    apperrors "loviary.app/backend/pkg/errors"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
    authService *appAuth.Service
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *appAuth.Service) *AuthHandler {
    return &AuthHandler{authService: authService}
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
// POST /api/v1/auth/register
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
        Password: req.Password,
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
// POST /api/v1/auth/login
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
// POST /api/v1/auth/refresh
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
            "access_token": result.AccessToken,
            "refresh_token": result.RefreshToken,
            "expires_in":   result.ExpiresIn,
        },
    })
}

// Logout handles user logout
// POST /api/v1/auth/logout
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
// POST /api/v1/auth/logout-all
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
