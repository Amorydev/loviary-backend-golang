package handlers

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"

    appUsers "loviary.app/backend/internal/application/users"
    "loviary.app/backend/internal/interfaces/http/dto"
    "loviary.app/backend/internal/interfaces/http/middleware"
    "loviary.app/backend/internal/domain/shared"
    apperrors "loviary.app/backend/pkg/errors"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
    userService *appUsers.Service
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *appUsers.Service) *UserHandler {
    return &UserHandler{userService: userService}
}

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
    Username    string     `json:"username" binding:"required,min=3,max=50,alphanum"`
    Email       string     `json:"email" binding:"required,email,max=100"`
    Password    string     `json:"password" binding:"required,min=8,max=100"`
    FirstName   *string    `json:"first_name" binding:"max=50"`
    LastName    *string    `json:"last_name" binding:"max=50"`
    DateOfBirth *time.Time `json:"date_of_birth"`
    Gender      *shared.Gender `json:"gender" binding:"omitempty,oneof=male female other prefer_not"`
    Language    string     `json:"language" binding:"max=10"`
}

// UpdateUserRequest represents update user request
type UpdateUserRequest struct {
    FirstName     *string    `json:"first_name" binding:"max=50"`
    LastName      *string    `json:"last_name" binding:"max=50"`
    DateOfBirth   *time.Time `json:"date_of_birth"`
    Gender        *shared.Gender `json:"gender" binding:"omitempty,oneof=male female other prefer_not"`
    Language      string     `json:"language" binding:"max=10"`
    AvatarURL     *string    `json:"avatar_url"`
}

// UpdateDeviceRequest represents update device request
type UpdateDeviceRequest struct {
    FCMToken   string `json:"fcm_token" binding:"required"`
    Platform   string `json:"platform" binding:"required,oneof=ios android web"`
    DeviceName string `json:"device_name"`
}

// CreateUser handles user registration
// POST /api/v1/users/register
func (h *UserHandler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", err.Error()))
        return
    }

    input := appUsers.CreateUserInput{
        Username:    req.Username,
        Email:       req.Email,
        Password:    req.Password,
        FirstName:   req.FirstName,
        LastName:    req.LastName,
        DateOfBirth: req.DateOfBirth,
        Gender:      req.Gender,
        Language:    req.Language,
    }

    user, err := h.userService.Create(c.Request.Context(), input)
    if err != nil {
        c.Error(err)
        return
    }

    c.JSON(http.StatusCreated, dto.UserToResponse(user))
}

// GetUser retrieves a user by ID
// GET /api/v1/users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", "Invalid user ID"))
        return
    }

    user, err := h.userService.GetByID(c.Request.Context(), id)
    if err != nil {
        c.Error(err)
        return
    }

    c.JSON(http.StatusOK, dto.UserToResponse(user))
}

// GetMyProfile returns the authenticated user's profile
// GET /api/v1/users/me
func (h *UserHandler) GetMyProfile(c *gin.Context) {
    userID, exists := middleware.GetUserID(c)
    if !exists {
        c.JSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
        return
    }

    user, err := h.userService.GetByID(c.Request.Context(), userID)
    if err != nil {
        c.Error(err)
        return
    }

    c.JSON(http.StatusOK, dto.UserToResponse(user))
}

// UpdateUser updates user profile
// PATCH /api/v1/users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", "Invalid user ID"))
        return
    }

    var req UpdateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", err.Error()))
        return
    }

    // Get existing user
    existingUser, err := h.userService.GetByID(c.Request.Context(), id)
    if err != nil {
        c.Error(err)
        return
    }

    // Update fields
    existingUser.FirstName = req.FirstName
    existingUser.LastName = req.LastName
    existingUser.DateOfBirth = req.DateOfBirth
    existingUser.Gender = req.Gender
    existingUser.Language = req.Language
    existingUser.AvatarURL = req.AvatarURL
    existingUser.UpdatedAt = time.Now()

    // TODO: Add Update method to service
    c.JSON(http.StatusOK, dto.UserToResponse(existingUser))
}

// UpdateMyProfile updates the authenticated user's profile
// PATCH /api/v1/users/me
func (h *UserHandler) UpdateMyProfile(c *gin.Context) {
    userID, exists := middleware.GetUserID(c)
    if !exists {
        c.JSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
        return
    }

    var req UpdateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", err.Error()))
        return
    }

    user, err := h.userService.GetByID(c.Request.Context(), userID)
    if err != nil {
        c.Error(err)
        return
    }

    // Update fields
    if req.FirstName != nil {
        user.FirstName = req.FirstName
    }
    if req.LastName != nil {
        user.LastName = req.LastName
    }
    if req.DateOfBirth != nil {
        user.DateOfBirth = req.DateOfBirth
    }
    if req.Gender != nil {
        user.Gender = req.Gender
    }
    if req.Language != "" {
        user.Language = req.Language
    }
    if req.AvatarURL != nil {
        user.AvatarURL = req.AvatarURL
    }
    user.UpdatedAt = time.Now()

    // TODO: Add Update method to service
    c.JSON(http.StatusOK, dto.UserToResponse(user))
}

// UpdateMyDevice updates the authenticated user's device
// PUT /api/v1/users/me/device
func (h *UserHandler) UpdateMyDevice(c *gin.Context) {
    userID, exists := middleware.GetUserID(c)
    if !exists {
        c.JSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
        return
    }

    var req UpdateDeviceRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", err.Error()))
        return
    }

    // TODO: Implement UpdateDevice in service
    _ = userID
    _ = req

    c.JSON(http.StatusOK, gin.H{"success": true})
}
