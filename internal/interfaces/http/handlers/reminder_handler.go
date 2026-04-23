package handlers

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"

    appReminders "loviary.app/backend/internal/application/reminders"
    "loviary.app/backend/internal/domain/reminders"
    "loviary.app/backend/internal/interfaces/http/dto"
    "loviary.app/backend/internal/interfaces/http/middleware"
    apperrors "loviary.app/backend/pkg/errors"
)

// ReminderHandler handles reminder-related HTTP requests
type ReminderHandler struct {
    reminderService *appReminders.Service
}

// NewReminderHandler creates a new reminder handler
func NewReminderHandler(reminderService *appReminders.Service) *ReminderHandler {
    return &ReminderHandler{reminderService: reminderService}
}

// CreateReminderRequest represents create reminder request
type CreateReminderRequest struct {
    Title          string                 `json:"title" binding:"required"`
    Description    *string                `json:"description"`
    ReminderTime   time.Time              `json:"reminder_time" binding:"required"`
    Recurrence     reminders.RecurrenceType `json:"recurrence" binding:"required,oneof=none daily weekly monthly custom"`
    RecurrenceValue *int                  `json:"recurrence_value"`
}

// UpdateReminderRequest represents update reminder request
type UpdateReminderRequest struct {
    Title          *string                `json:"title"`
    Description    *string                `json:"description"`
    ReminderTime   *time.Time             `json:"reminder_time"`
    Recurrence     *reminders.RecurrenceType `json:"recurrence" binding:"omitempty,oneof=none daily weekly monthly custom"`
    RecurrenceValue *int                  `json:"recurrence_value"`
    IsActive       *bool                  `json:"is_active"`
}

// CreateReminder creates a new reminder
// POST /api/v1/reminders
func (h *ReminderHandler) CreateReminder(c *gin.Context) {
    userID, exists := middleware.GetUserID(c)
    if !exists {
        c.JSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
        return
    }

    var req CreateReminderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", err.Error()))
        return
    }

    reminder, err := h.reminderService.Create(c.Request.Context(), appReminders.CreateReminderInput{
        UserID:         userID,
        Title:          req.Title,
        Description:    req.Description,
        ReminderTime:   req.ReminderTime,
        Recurrence:     req.Recurrence,
        RecurrenceValue: req.RecurrenceValue,
    })
    if err != nil {
        c.Error(err)
        return
    }

    c.JSON(http.StatusCreated, dto.ReminderToResponse(reminder))
}

// GetReminder retrieves a reminder by ID
// GET /api/v1/reminders/:id
func (h *ReminderHandler) GetReminder(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", "Invalid reminder ID"))
        return
    }

    reminder, err := h.reminderService.GetReminder(c.Request.Context(), id)
    if err != nil {
        c.Error(err)
        return
    }

    c.JSON(http.StatusOK, dto.ReminderToResponse(reminder))
}

// GetReminders retrieves all reminders for the authenticated user
// GET /api/v1/reminders
func (h *ReminderHandler) GetReminders(c *gin.Context) {
    userID, exists := middleware.GetUserID(c)
    if !exists {
        c.JSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
        return
    }

    reminders, err := h.reminderService.GetRemindersByUser(c.Request.Context(), userID)
    if err != nil {
        c.Error(err)
        return
    }

    response := make([]dto.ReminderResponse, len(reminders))
    for i, reminder := range reminders {
        response[i] = dto.ReminderToResponse(&reminder)
    }

    c.JSON(http.StatusOK, gin.H{
        "reminders": response,
        "count":     len(response),
    })
}

// UpdateReminder updates a reminder
// PATCH /api/v1/reminders/:id
func (h *ReminderHandler) UpdateReminder(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", "Invalid reminder ID"))
        return
    }

    var req UpdateReminderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", err.Error()))
        return
    }

    reminder, err := h.reminderService.Update(c.Request.Context(), id, appReminders.UpdateReminderInput{
        Title:          req.Title,
        Description:    req.Description,
        ReminderTime:   req.ReminderTime,
        Recurrence:     req.Recurrence,
        RecurrenceValue: req.RecurrenceValue,
        IsActive:       req.IsActive,
    })
    if err != nil {
        c.Error(err)
        return
    }

    c.JSON(http.StatusOK, dto.ReminderToResponse(reminder))
}

// DeleteReminder removes a reminder
// DELETE /api/v1/reminders/:id
func (h *ReminderHandler) DeleteReminder(c *gin.Context) {
    idStr := c.Param("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", "Invalid reminder ID"))
        return
    }

    if err := h.reminderService.DeleteReminder(c.Request.Context(), id); err != nil {
        c.Error(err)
        return
    }

    c.JSON(http.StatusOK, gin.H{"success": true})
}
