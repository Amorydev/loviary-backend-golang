package handlers

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"

    appMoods "loviary.app/backend/internal/application/moods"
    "loviary.app/backend/internal/domain/moods"
    "loviary.app/backend/internal/interfaces/http/dto"
    "loviary.app/backend/internal/interfaces/http/middleware"
    apperrors "loviary.app/backend/pkg/errors"
)

// MoodHandler handles mood-related HTTP requests
type MoodHandler struct {
    moodService *appMoods.Service
}

// NewMoodHandler creates a new mood handler
func NewMoodHandler(moodService *appMoods.Service) *MoodHandler {
    return &MoodHandler{moodService: moodService}
}

// CreateMoodRequest represents create mood request
type CreateMoodRequest struct {
    Date      time.Time             `json:"date" binding:"required"`
    MoodType  moods.MoodType        `json:"mood_type" binding:"required,oneof=happy sad angry anxious calm excited tired love stressed grateful"`
    Intensity int                   `json:"intensity" binding:"required,min=1,max=10"`
    Note      *string               `json:"note"`
    IsShared  bool                  `json:"is_shared"`
}

// UpdateMoodRequest represents update mood request
type UpdateMoodRequest struct {
    MoodType  *moods.MoodType       `json:"mood_type" binding:"omitempty,oneof=happy sad angry anxious calm excited tired love stressed grateful"`
    Intensity *int                  `json:"intensity" binding:"omitempty,min=1,max=10"`
    Note      *string               `json:"note"`
    IsShared  *bool                 `json:"is_shared"`
}

// CreateMood creates a new mood entry
// POST /api/v1/moods
func (h *MoodHandler) CreateMood(c *gin.Context) {
    userID, exists := middleware.GetUserID(c)
    if !exists {
        c.JSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
        return
    }

    var req CreateMoodRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", err.Error()))
        return
    }

    mood, err := h.moodService.Create(c.Request.Context(), appMoods.CreateMoodInput{
        UserID:    userID,
        Date:      req.Date,
        MoodType:  req.MoodType,
        Intensity: req.Intensity,
        Note:      req.Note,
        IsShared:  req.IsShared,
    })
    if err != nil {
        c.Error(err)
        return
    }

    c.JSON(http.StatusCreated, dto.MoodToResponse(mood))
}

// GetMood retrieves a mood by ID
// GET /api/v1/moods/:id
func (h *MoodHandler) GetMood(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", "Invalid mood ID"))
        return
    }

    mood, err := h.moodService.GetMood(c.Request.Context(), id)
    if err != nil {
        c.Error(err)
        return
    }

    c.JSON(http.StatusOK, dto.MoodToResponse(mood))
}

// UpdateMood updates a mood
// PATCH /api/v1/moods/:id
func (h *MoodHandler) UpdateMood(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", "Invalid mood ID"))
        return
    }

    var req UpdateMoodRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", err.Error()))
        return
    }

    mood, err := h.moodService.Update(c.Request.Context(), id, appMoods.UpdateMoodInput{
        MoodType:  req.MoodType,
        Intensity: req.Intensity,
        Note:      req.Note,
        IsShared:  req.IsShared,
    })
    if err != nil {
        c.Error(err)
        return
    }

    c.JSON(http.StatusOK, dto.MoodToResponse(mood))
}

// GetTodaysMood retrieves today's mood
// GET /api/v1/moods/today
func (h *MoodHandler) GetTodaysMood(c *gin.Context) {
    userID, exists := middleware.GetUserID(c)
    if !exists {
        c.JSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
        return
    }

    mood, err := h.moodService.GetTodaysMood(c.Request.Context(), userID)
    if err != nil {
        c.Error(err)
        return
    }

    if mood == nil {
        c.JSON(http.StatusNotFound, apperrors.New("NO_MOOD_TODAY", "No mood logged today"))
        return
    }

    c.JSON(http.StatusOK, dto.MoodToResponse(mood))
}

// GetMoodHistory retrieves mood history
// GET /api/v1/moods/history?start_date=...&end_date=...
func (h *MoodHandler) GetMoodHistory(c *gin.Context) {
    userID, exists := middleware.GetUserID(c)
    if !exists {
        c.JSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
        return
    }

    // Parse query parameters
    startDateStr := c.Query("start_date")
    endDateStr := c.Query("end_date")

    var startDate, endDate time.Time
    if startDateStr != "" {
        parsedStart, err := time.Parse(time.RFC3339, startDateStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", "Invalid start_date format"))
            return
        }
        startDate = parsedStart
    }
    if endDateStr != "" {
        parsedEnd, err := time.Parse(time.RFC3339, endDateStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", "Invalid end_date format"))
            return
        }
        endDate = parsedEnd
    }

    moodList, err := h.moodService.GetMoodsByUser(c.Request.Context(), userID, startDate, endDate)
    if err != nil {
        c.Error(err)
        return
    }

    response := make([]dto.MoodResponse, len(moodList))
    for i, mood := range moodList {
        response[i] = dto.MoodToResponse(&mood)
    }

    c.JSON(http.StatusOK, gin.H{
        "moods": response,
        "count": len(response),
    })
}

// DeleteMood removes a mood
// DELETE /api/v1/moods/:id
func (h *MoodHandler) DeleteMood(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", "Invalid mood ID"))
        return
    }

    // Verify ownership (optional - could be done at DB level)
    // For now, just delete
    if err := h.moodService.DeleteMood(c.Request.Context(), id); err != nil {
        c.Error(err)
        return
    }

    c.JSON(http.StatusOK, gin.H{"success": true})
}
