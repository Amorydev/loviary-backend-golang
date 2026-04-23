package reminders

import (
    "context"
    "errors"
    "time"

    "github.com/google/uuid"

    "loviary.app/backend/internal/domain/reminders"
    apperrors "loviary.app/backend/pkg/errors"
)

// Repository defines the interface for reminder persistence
type Repository interface {
    Create(ctx context.Context, reminder *reminders.Reminder) error
    GetByID(ctx context.Context, id uuid.UUID) (*reminders.Reminder, error)
    GetByUserID(ctx context.Context, userID uuid.UUID) ([]reminders.Reminder, error)
    Update(ctx context.Context, reminder *reminders.Reminder) error
    Delete(ctx context.Context, id uuid.UUID) error
    UpdateLastSent(ctx context.Context, id uuid.UUID, sentAt time.Time) error
    GetDueReminders(ctx context.Context, before time.Time) ([]reminders.Reminder, error)
}

// Service handles reminder business logic
type Service struct {
    repo Repository
}

// NewService creates a new reminder service
func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}

// CreateReminderInput represents input for creating a reminder
type CreateReminderInput struct {
    UserID         uuid.UUID
    Title          string
    Description    *string
    ReminderTime   time.Time
    Recurrence     reminders.RecurrenceType
    RecurrenceValue *int
}

// Create creates a new reminder
func (s *Service) Create(ctx context.Context, input CreateReminderInput) (*reminders.Reminder, error) {
    // Validate recurrence type
    if !input.Recurrence.IsValid() {
        return nil, apperrors.New("INVALID_RECURRENCE", "Invalid recurrence type")
    }

    // Validate recurrence value if provided
    if input.RecurrenceValue != nil {
        if *input.RecurrenceValue <= 0 {
            return nil, apperrors.New("INVALID_RECURRENCE_VALUE", "Recurrence value must be positive")
        }
    }

    now := time.Now()
    reminder := &reminders.Reminder{
        ID:              uuid.New(),
        UserID:          input.UserID,
        Title:           input.Title,
        Description:     input.Description,
        ReminderTime:    input.ReminderTime,
        Recurrence:      input.Recurrence,
        RecurrenceValue: input.RecurrenceValue,
        IsActive:        true,
        CreatedAt:       now,
        UpdatedAt:       now,
    }

    if err := s.repo.Create(ctx, reminder); err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to create reminder")
    }

    return reminder, nil
}

// UpdateReminderInput represents input for updating a reminder
type UpdateReminderInput struct {
    Title          *string
    Description    *string
    ReminderTime   *time.Time
    Recurrence     *reminders.RecurrenceType
    RecurrenceValue *int
    IsActive       *bool
}

// Update updates an existing reminder
func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateReminderInput) (*reminders.Reminder, error) {
    reminder, err := s.repo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, apperrors.New("NOT_FOUND", "")) {
            return nil, apperrors.New("REMINDER_NOT_FOUND", "Reminder not found")
        }
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get reminder")
    }

    // Update fields if provided
    if input.Title != nil {
        reminder.Title = *input.Title
    }
    if input.Description != nil {
        reminder.Description = input.Description
    }
    if input.ReminderTime != nil {
        reminder.ReminderTime = *input.ReminderTime
    }
    if input.Recurrence != nil {
        if !(*input.Recurrence).IsValid() {
            return nil, apperrors.New("INVALID_RECURRENCE", "Invalid recurrence type")
        }
        reminder.Recurrence = *input.Recurrence
    }
    if input.RecurrenceValue != nil {
        if *input.RecurrenceValue <= 0 {
            return nil, apperrors.New("INVALID_RECURRENCE_VALUE", "Recurrence value must be positive")
        }
        reminder.RecurrenceValue = input.RecurrenceValue
    }
    if input.IsActive != nil {
        reminder.IsActive = *input.IsActive
    }

    reminder.UpdatedAt = time.Now()

    if err := s.repo.Update(ctx, reminder); err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to update reminder")
    }

    return reminder, nil
}

// GetReminder retrieves a reminder by ID
func (s *Service) GetReminder(ctx context.Context, id uuid.UUID) (*reminders.Reminder, error) {
    reminder, err := s.repo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, apperrors.New("NOT_FOUND", "")) {
            return nil, apperrors.New("REMINDER_NOT_FOUND", "Reminder not found")
        }
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get reminder")
    }
    return reminder, nil
}

// GetRemindersByUser retrieves all reminders for a user
func (s *Service) GetRemindersByUser(ctx context.Context, userID uuid.UUID) ([]reminders.Reminder, error) {
    reminders, err := s.repo.GetByUserID(ctx, userID)
    if err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get reminders")
    }
    return reminders, nil
}

// DeleteReminder removes a reminder
func (s *Service) DeleteReminder(ctx context.Context, id uuid.UUID) error {
    // First verify the reminder exists
    _, err := s.repo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, apperrors.ReminderNotFound) {
            return apperrors.New("REMINDER_NOT_FOUND", "Reminder not found")
        }
        return apperrors.New("INTERNAL_ERROR", "Failed to get reminder")
    }

    if err := s.repo.Delete(ctx, id); err != nil {
        return apperrors.New("INTERNAL_ERROR", "Failed to delete reminder")
    }
    return nil
}

// MarkAsSent updates the last_sent_at timestamp
func (s *Service) MarkAsSent(ctx context.Context, id uuid.UUID) error {
    if err := s.repo.UpdateLastSent(ctx, id, time.Now()); err != nil {
        return apperrors.New("INTERNAL_ERROR", "Failed to mark reminder as sent")
    }
    return nil
}

// GetDueReminders retrieves reminders that are due
func (s *Service) GetDueReminders(ctx context.Context) ([]reminders.Reminder, error) {
    due, err := s.repo.GetDueReminders(ctx, time.Now())
    if err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get due reminders")
    }
    return due, nil
}
