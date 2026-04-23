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
// GET /api/v1/streaks/me
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
// GET /api/v1/streaks/:activity_type
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
// POST /api/v1/streaks/:activity_type/record
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
