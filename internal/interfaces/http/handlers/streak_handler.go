package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appStreaks "loviary.app/backend/internal/application/streaks"
	"loviary.app/backend/internal/interfaces/http/dto"
	"loviary.app/backend/internal/interfaces/http/middleware"
	apperrors "loviary.app/backend/pkg/errors"
)

// StreakHandler handles streak-related HTTP requests
type StreakHandler struct {
	streakService *appStreaks.Service
}

// NewStreakHandler creates a new streak handler
func NewStreakHandler(streakService *appStreaks.Service) *StreakHandler {
	return &StreakHandler{streakService: streakService}
}

// GetMyStreaks retrieves the authenticated user's streaks
// @Summary Get all streaks
// @Description Get all activity streaks for the current user
// @Tags streaks
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success  200  {object}  handlers.StreakListResponse "List of streaks with count"
// @Failure  401  {object}  handlers.ErrorResponse "Not authenticated"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /streaks/me [get]
func (h *StreakHandler) GetMyStreaks(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
		return
	}

	streakList, err := h.streakService.GetAllStreaks(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}

	response := make([]dto.StreakResponse, len(streakList))
	for i, streak := range streakList {
		response[i] = dto.StreakToResponse(&streak)
	}

	c.JSON(http.StatusOK, gin.H{
		"streaks": response,
		"count":   len(response),
	})
}

// GetStreak retrieves a specific streak
// @Summary Get streak
// @Description Get streak for a specific activity type
// @Tags streaks
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param   activity_type  path  string  true  "Activity type (e.g., date_night, exercise)"
// @Success  200  {object}  dto.StreakResponse "Streak data"
// @Failure  400  {object}  handlers.ErrorResponse "Activity type required"
// @Failure  401  {object}  handlers.ErrorResponse "Not authenticated"
// @Failure  404  {object}  handlers.ErrorResponse "Streak not found"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /streaks/{activity_type} [get]
func (h *StreakHandler) GetStreak(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
		return
	}

	activityType := c.Param("activity_type")
	if activityType == "" {
		c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", "Activity type is required"))
		return
	}

	streak, err := h.streakService.GetStreak(c.Request.Context(), userID, activityType)
	if err != nil {
		c.Error(err)
		return
	}

	if streak == nil {
		c.JSON(http.StatusNotFound, apperrors.New("STREAK_NOT_FOUND", "No streak found for this activity"))
		return
	}

	c.JSON(http.StatusOK, dto.StreakToResponse(streak))
}

// RecordActivity records activity and updates streak
// @Summary Record activity
// @Description Record an activity occurrence and update streak counters
// @Tags streaks
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param   activity_type  path  string  true  "Activity type (e.g., date_night, exercise)"
// @Success  200  {object}  dto.StreakResponse "Updated streak"
// @Failure  400  {object}  handlers.ErrorResponse "Activity type required"
// @Failure  401  {object}  handlers.ErrorResponse "Not authenticated"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /streaks/{activity_type}/record [post]
func (h *StreakHandler) RecordActivity(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
		return
	}

	activityType := c.Param("activity_type")
	if activityType == "" {
		c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", "Activity type is required"))
		return
	}

	streak, err := h.streakService.RecordActivity(c.Request.Context(), userID, activityType)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.StreakToResponse(streak))
}
