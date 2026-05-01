package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	appCouples "loviary.app/backend/internal/application/couples"
	appUsers "loviary.app/backend/internal/application/users"
	"loviary.app/backend/internal/domain/shared"
	"loviary.app/backend/internal/interfaces/http/dto"
	"loviary.app/backend/internal/interfaces/http/middleware"
	apperrors "loviary.app/backend/pkg/errors"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService   *appUsers.Service
	coupleService *appCouples.Service
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *appUsers.Service, coupleService *appCouples.Service) *UserHandler {
	return &UserHandler{
		userService:   userService,
		coupleService: coupleService,
	}
}

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	Username    string         `json:"username" binding:"required,min=3,max=50,alphanum"`
	Email       string         `json:"email" binding:"required,email,max=100"`
	Password    string         `json:"password" binding:"required,min=8,max=100"`
	FirstName   *string        `json:"first_name" binding:"max=50"`
	LastName    *string        `json:"last_name" binding:"max=50"`
	DateOfBirth *time.Time     `json:"date_of_birth"`
	Gender      *shared.Gender `json:"gender" binding:"omitempty,oneof=male female other prefer_not"`
	Language    string         `json:"language" binding:"max=10"`
}

// UpdateUserRequest represents update user request
type UpdateUserRequest struct {
	FirstName   *string        `json:"first_name" binding:"max=50"`
	LastName    *string        `json:"last_name" binding:"max=50"`
	DateOfBirth *time.Time     `json:"date_of_birth"`
	Gender      *shared.Gender `json:"gender" binding:"omitempty,oneof=male female other prefer_not"`
	Language    string         `json:"language" binding:"max=10"`
	AvatarURL   *string        `json:"avatar_url"`
}

// UpdateDeviceRequest represents update device request
type UpdateDeviceRequest struct {
	FCMToken   string `json:"fcm_token" binding:"required"`
	Platform   string `json:"platform" binding:"required,oneof=ios android web"`
	DeviceName string `json:"device_name"`
}

// CreateUser handles user registration
// @Summary Create user
// @Description Create a new user account (alternative endpoint)
// @Tags users
// @Accept  json
// @Produce  json
// @Param   request  body  handlers.CreateUserRequest  true  "User creation request"
// @Success  201  {object}  dto.UserResponse "User created successfully"
// @Failure  400  {object}  handlers.ErrorResponse "Invalid input"
// @Failure  409  {object}  handlers.ErrorResponse "Duplicate email or username"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /users/register [post]
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

	c.JSON(http.StatusCreated, dto.UserToResponse(user, false))
}

// GetUser retrieves a user by ID
// @Summary Get user by ID
// @Description Get user information by user ID
// @Tags users
// @Accept  json
// @Produce  json
// @Param   id  path  string  true  "User ID (UUID)"
// @Success  200  {object}  dto.UserResponse "User found"
// @Failure  400  {object}  handlers.ErrorResponse "Invalid user ID"
// @Failure  404  {object}  handlers.ErrorResponse "User not found"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /users/{id} [get]
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

	c.JSON(http.StatusOK, dto.UserToResponse(user, false))
}

// GetMyProfile returns the authenticated user's profile
// @Summary Get my profile
// @Description Get current authenticated user's profile information
// @Tags users
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success  200  {object}  dto.UserResponse "User profile"
// @Failure  401  {object}  handlers.ErrorResponse "Not authenticated"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /users/me [get]
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

	// Resolve couple status for unified model
	couple, _ := h.coupleService.GetActiveByUserID(c.Request.Context(), userID)
	hasCouple := couple != nil

	c.JSON(http.StatusOK, dto.UserToResponse(user, hasCouple))
}

// UpdateUser updates user profile
// @Summary Update user
// @Description Update user profile information by ID
// @Tags users
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param   id  path  string  true  "User ID (UUID)"
// @Param   request  body  handlers.UpdateUserRequest  true  "User update data"
// @Success  200  {object}  dto.UserResponse "Updated user data"
// @Failure  400  {object}  handlers.ErrorResponse "Invalid input or user ID"
// @Failure  401  {object}  handlers.ErrorResponse "Not authenticated"
// @Failure  404  {object}  handlers.ErrorResponse "User not found"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /users/{id} [patch]
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
	c.JSON(http.StatusOK, dto.UserToResponse(existingUser, false))
}

// UpdateMyProfile updates the authenticated user's profile
// @Summary Update my profile
// @Description Update current authenticated user's profile
// @Tags users
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param   request  body  handlers.UpdateUserRequest  true  "User update data"
// @Success  200  {object}  dto.UserResponse "Updated user data"
// @Failure  400  {object}  handlers.ErrorResponse "Invalid input"
// @Failure  401  {object}  handlers.ErrorResponse "Not authenticated"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /users/me [patch]
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
	couple, _ := h.coupleService.GetActiveByUserID(c.Request.Context(), userID)
	hasCouple := couple != nil
	c.JSON(http.StatusOK, dto.UserToResponse(user, hasCouple))
}

// UpdateMyDevice updates the authenticated user's device
// @Summary Update my device
// @Description Update device information (FCM token, platform) for push notifications
// @Tags users
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param   request  body  handlers.UpdateDeviceRequest  true  "Device info"
// @Success  200  {object}  handlers.SuccessResponse "Device updated"
// @Failure  400  {object}  handlers.ErrorResponse "Invalid input"
// @Failure  401  {object}  handlers.ErrorResponse "Not authenticated"
// @Router   /users/me/device [put]
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
