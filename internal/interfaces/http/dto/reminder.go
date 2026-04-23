package dto

import (
    "time"

    "github.com/google/uuid"

    "loviary.app/backend/internal/domain/reminders"
)

// ReminderResponse represents reminder response DTO
type ReminderResponse struct {
    ID              uuid.UUID             `json:"id"`
    UserID          uuid.UUID             `json:"user_id"`
    Title           string                `json:"title"`
    Description     *string               `json:"description,omitempty"`
    ReminderTime    time.Time             `json:"reminder_time"`
    Recurrence      reminders.RecurrenceType `json:"recurrence"`
    RecurrenceValue *int                  `json:"recurrence_value,omitempty"`
    IsActive        bool                  `json:"is_active"`
    LastSentAt      *time.Time            `json:"last_sent_at,omitempty"`
    CreatedAt       time.Time             `json:"created_at"`
    UpdatedAt       time.Time             `json:"updated_at"`
}

// ReminderToResponse converts a Reminder to response DTO
func ReminderToResponse(reminder *reminders.Reminder) ReminderResponse {
    return ReminderResponse{
        ID:              reminder.ID,
        UserID:          reminder.UserID,
        Title:           reminder.Title,
        Description:     reminder.Description,
        ReminderTime:    reminder.ReminderTime,
        Recurrence:      reminder.Recurrence,
        RecurrenceValue: reminder.RecurrenceValue,
        IsActive:        reminder.IsActive,
        LastSentAt:      reminder.LastSentAt,
        CreatedAt:       reminder.CreatedAt,
        UpdatedAt:       reminder.UpdatedAt,
    }
}
