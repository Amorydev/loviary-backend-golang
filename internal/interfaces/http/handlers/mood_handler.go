package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	appMoods "loviary.app/backend/internal/application/moods"
	appStreaks "loviary.app/backend/internal/application/streaks"
	"loviary.app/backend/internal/domain/moods"
	"loviary.app/backend/internal/interfaces/http/dto"
	"loviary.app/backend/internal/interfaces/http/middleware"
	apperrors "loviary.app/backend/pkg/errors"
)

// MoodHandler handles mood-related HTTP requests.
type MoodHandler struct {
	moodService   *appMoods.Service
	streakService *appStreaks.Service
}

// NewMoodHandler creates a new mood handler.
func NewMoodHandler(moodService *appMoods.Service, streakService *appStreaks.Service) *MoodHandler {
	return &MoodHandler{
		moodService:   moodService,
		streakService: streakService,
	}
}

// CreateMoodRequest represents create mood request
type CreateMoodRequest struct {
	Date      time.Time      `json:"date" binding:"required"`
	MoodType  moods.MoodType `json:"mood_type" binding:"required,oneof=happy sad angry anxious calm excited tired love stressed grateful"`
	Intensity int            `json:"intensity" binding:"required,min=1,max=10"`
	Note      *string        `json:"note"`
	IsShared  bool           `json:"is_shared"`
}

// UpdateMoodRequest represents update mood request
type UpdateMoodRequest struct {
	MoodType  *moods.MoodType `json:"mood_type" binding:"omitempty,oneof=happy sad angry anxious calm excited tired love stressed grateful"`
	Intensity *int            `json:"intensity" binding:"omitempty,min=1,max=10"`
	Note      *string         `json:"note"`
	IsShared  *bool           `json:"is_shared"`
}

// CreateMood creates a new mood entry
// @Summary Create mood
// @Description Log a new mood entry with type and intensity
// @Tags moods
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param   request  body  handlers.CreateMoodRequest  true  "Mood data"
// @Success  201  {object}  dto.MoodResponse "Mood created"
// @Failure  400  {object}  handlers.ErrorResponse "Invalid input"
// @Failure  401  {object}  handlers.ErrorResponse "Not authenticated"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /moods [post]
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

	// Trigger streak log asynchronously (fire-and-forget).
	// coupleID is embedded in JWT claims via middleware.
	if coupleIDPtr, ok := middleware.GetCoupleID(c); ok && coupleIDPtr != nil {
		coupleID := *coupleIDPtr
		go func() {
			if err := h.streakService.RecordLog(context.Background(), coupleID, userID, "mood_log"); err != nil {
				slog.Error("failed to record streak log after mood creation",
					"user_id", userID,
					"couple_id", coupleID,
					"error", err,
				)
			}
		}()
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": dto.MoodToResponse(mood)})
}

// GetMood retrieves a mood by ID
// @Summary Get mood
// @Description Get mood entry by ID
// @Tags moods
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param   id  path  string  true  "Mood ID (UUID)"
// @Success  200  {object}  dto.MoodResponse "Mood data"
// @Failure  400  {object}  handlers.ErrorResponse "Invalid mood ID"
// @Failure  401  {object}  handlers.ErrorResponse "Not authenticated"
// @Failure  404  {object}  handlers.ErrorResponse "Mood not found"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /moods/{id} [get]
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
// @Summary Update mood
// @Description Update an existing mood entry
// @Tags moods
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param   id  path  string  true  "Mood ID (UUID)"
// @Param   request  body  handlers.UpdateMoodRequest  true  "Updated mood data"
// @Success  200  {object}  dto.MoodResponse "Updated mood"
// @Failure  400  {object}  handlers.ErrorResponse "Invalid input"
// @Failure  401  {object}  handlers.ErrorResponse "Not authenticated"
// @Failure  404  {object}  handlers.ErrorResponse "Mood not found"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /moods/{id} [patch]
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
// @Summary Get today's mood
// @Description Get mood logged for today
// @Tags moods
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success  200  {object}  dto.MoodResponse "Today's mood"
// @Failure  401  {object}  handlers.ErrorResponse "Not authenticated"
// @Failure  404  {object}  handlers.ErrorResponse "No mood logged today"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /moods/today [get]
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
// @Summary Get mood history
// @Description Get mood history with optional date range filtering
// @Tags moods
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param   start_date  query  string  false  "Start date (RFC3339 format)"
// @Param   end_date    query  string  false  "End date (RFC3339 format)"
// @Success  200  {object}  handlers.MoodListResponse "Mood history with count"
// @Failure  400  {object}  handlers.ErrorResponse "Invalid date format"
// @Failure  401  {object}  handlers.ErrorResponse "Not authenticated"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /moods/history [get]
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
// @Summary Delete mood
// @Description Delete a mood entry by ID
// @Tags moods
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param   id  path  string  true  "Mood ID (UUID)"
// @Success  200  {object}  handlers.SuccessResponse "Mood deleted"
// @Failure  400  {object}  handlers.ErrorResponse "Invalid mood ID"
// @Failure  401  {object}  handlers.ErrorResponse "Not authenticated"
// @Failure  404  {object}  handlers.ErrorResponse "Mood not found"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /moods/{id} [delete]
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
