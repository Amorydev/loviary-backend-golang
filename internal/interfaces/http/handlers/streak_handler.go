package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appStreaks "loviary.app/backend/internal/application/streaks"
	"loviary.app/backend/internal/interfaces/http/dto"
	"loviary.app/backend/internal/interfaces/http/middleware"
	apperrors "loviary.app/backend/pkg/errors"
)

// StreakHandler handles streak-related HTTP requests.
// All streak operations are per-couple.
type StreakHandler struct {
	streakService *appStreaks.Service
}

// NewStreakHandler creates a new streak handler.
func NewStreakHandler(streakService *appStreaks.Service) *StreakHandler {
	return &StreakHandler{streakService: streakService}
}

// GetMyStreaks retrieves the couple's streaks for the authenticated user.
// @Summary      Get couple streaks
// @Description  Get all activity streaks for the current couple
// @Tags         streaks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  handlers.ErrorResponse
// @Failure      500  {object}  handlers.ErrorResponse
// @Router       /streaks/me [get]
func (h *StreakHandler) GetMyStreaks(c *gin.Context) {
	coupleIDPtr, ok := middleware.GetCoupleID(c)
	if !ok || coupleIDPtr == nil {
		c.JSON(http.StatusUnprocessableEntity, apperrors.New("NO_COUPLE", "You are not part of an active couple"))
		return
	}

	streakList, err := h.streakService.GetAllStreaks(c.Request.Context(), *coupleIDPtr)
	if err != nil {
		c.Error(err)
		return
	}

	response := make([]dto.StreakResponse, len(streakList))
	for i := range streakList {
		response[i] = dto.StreakToResponse(&streakList[i])
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"streaks": response,
			"count":   len(response),
		},
	})
}

// GetStreak retrieves a specific streak by activity type for the couple.
// @Summary      Get streak
// @Description  Get streak for a specific activity type for the current couple
// @Tags         streaks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        activity_type  path  string  true  "Activity type (e.g., mood_log)"
// @Success      200  {object}  dto.StreakResponse
// @Failure      400  {object}  handlers.ErrorResponse
// @Failure      401  {object}  handlers.ErrorResponse
// @Failure      422  {object}  handlers.ErrorResponse
// @Failure      500  {object}  handlers.ErrorResponse
// @Router       /streaks/{activity_type} [get]
func (h *StreakHandler) GetStreak(c *gin.Context) {
	coupleIDPtr, ok := middleware.GetCoupleID(c)
	if !ok || coupleIDPtr == nil {
		c.JSON(http.StatusUnprocessableEntity, apperrors.New("NO_COUPLE", "You are not part of an active couple"))
		return
	}

	activityType := c.Param("activity_type")
	if activityType == "" {
		c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", "Activity type is required"))
		return
	}

	streak, err := h.streakService.GetStreakByCouple(c.Request.Context(), *coupleIDPtr, activityType)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": dto.StreakToResponse(streak)})
}

// RecordActivity is deprecated — streak logs are now triggered automatically from POST /moods.
// @Summary      Record activity (deprecated)
// @Description  Manually record an activity for the couple streak. Prefer POST /moods which triggers this automatically.
// @Tags         streaks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        activity_type  path  string  true  "Activity type (e.g., mood_log)"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  handlers.ErrorResponse
// @Failure      401  {object}  handlers.ErrorResponse
// @Failure      422  {object}  handlers.ErrorResponse
// @Failure      500  {object}  handlers.ErrorResponse
// @Router       /streaks/{activity_type}/record [post]
func (h *StreakHandler) RecordActivity(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
		return
	}

	coupleIDPtr, ok := middleware.GetCoupleID(c)
	if !ok || coupleIDPtr == nil {
		c.JSON(http.StatusUnprocessableEntity, apperrors.New("NO_COUPLE", "You are not part of an active couple"))
		return
	}

	activityType := c.Param("activity_type")
	if activityType == "" {
		c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", "Activity type is required"))
		return
	}

	if err := h.streakService.RecordLog(c.Request.Context(), *coupleIDPtr, userID, activityType); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Activity recorded"})
}
