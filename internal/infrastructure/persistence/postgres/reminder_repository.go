package persistence

import (
    "context"
    "database/sql"
    "errors"
    "time"

    "github.com/google/uuid"

    "loviary.app/backend/internal/domain/reminders"
    apperrors "loviary.app/backend/pkg/errors"
)

// ReminderRepository handles database operations for reminders
type ReminderRepository struct {
    db *sql.DB
}

// NewReminderRepository creates a new reminder repository
func NewReminderRepository(db *sql.DB) *ReminderRepository {
    return &ReminderRepository{db: db}
}

// Create inserts a new reminder
func (r *ReminderRepository) Create(ctx context.Context, reminder *reminders.Reminder) error {
    query := `
        INSERT INTO reminders (id, user_id, title, description, reminder_time, recurrence, recurrence_value, is_active, last_sent_at, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `
    _, err := r.db.ExecContext(ctx, query,
        reminder.ID,
        reminder.UserID,
        reminder.Title,
        reminder.Description,
        reminder.ReminderTime,
        reminder.Recurrence,
        reminder.RecurrenceValue,
        reminder.IsActive,
        reminder.LastSentAt,
        reminder.CreatedAt,
        reminder.UpdatedAt,
    )
    if err != nil {
        return apperrors.New("INTERNAL_ERROR", "Failed to create reminder")
    }
    return nil
}

// GetByID retrieves a reminder by ID
func (r *ReminderRepository) GetByID(ctx context.Context, id uuid.UUID) (*reminders.Reminder, error) {
    query := `
        SELECT id, user_id, title, description, reminder_time, recurrence, recurrence_value, is_active, last_sent_at, created_at, updated_at
        FROM reminders
        WHERE id = $1
    `
    var reminder reminders.Reminder
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &reminder.ID,
        &reminder.UserID,
        &reminder.Title,
        &reminder.Description,
        &reminder.ReminderTime,
        &reminder.Recurrence,
        &reminder.RecurrenceValue,
        &reminder.IsActive,
        &reminder.LastSentAt,
        &reminder.CreatedAt,
        &reminder.UpdatedAt,
    )
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, apperrors.ReminderNotFound
        }
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get reminder")
    }
    return &reminder, nil
}

// GetByUserID retrieves all reminders for a user
func (r *ReminderRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]reminders.Reminder, error) {
    query := `
        SELECT id, user_id, title, description, reminder_time, recurrence, recurrence_value, is_active, last_sent_at, created_at, updated_at
        FROM reminders
        WHERE user_id = $1
        ORDER BY reminder_time ASC
    `
    rows, err := r.db.QueryContext(ctx, query, userID)
    if err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to query reminders")
    }
    defer rows.Close()

    var reminderList []reminders.Reminder
    for rows.Next() {
        var reminder reminders.Reminder
        if err := rows.Scan(
            &reminder.ID,
            &reminder.UserID,
            &reminder.Title,
            &reminder.Description,
            &reminder.ReminderTime,
            &reminder.Recurrence,
            &reminder.RecurrenceValue,
            &reminder.IsActive,
            &reminder.LastSentAt,
            &reminder.CreatedAt,
            &reminder.UpdatedAt,
        ); err != nil {
            return nil, apperrors.New("INTERNAL_ERROR", "Failed to scan reminder row")
        }
        reminderList = append(reminderList, reminder)
    }
    if err := rows.Err(); err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to iterate reminder rows")
    }
    return reminderList, nil
}

// Update updates an existing reminder
func (r *ReminderRepository) Update(ctx context.Context, reminder *reminders.Reminder) error {
    query := `
        UPDATE reminders
        SET title = $2, description = $3, reminder_time = $4, recurrence = $5,
            recurrence_value = $6, is_active = $7, last_sent_at = $8, updated_at = $9
        WHERE id = $1
    `
    _, err := r.db.ExecContext(ctx, query,
        reminder.ID,
        reminder.Title,
        reminder.Description,
        reminder.ReminderTime,
        reminder.Recurrence,
        reminder.RecurrenceValue,
        reminder.IsActive,
        reminder.LastSentAt,
        reminder.UpdatedAt,
    )
    if err != nil {
        return apperrors.New("INTERNAL_ERROR", "Failed to update reminder")
    }
    return nil
}

// Delete removes a reminder
func (r *ReminderRepository) Delete(ctx context.Context, id uuid.UUID) error {
    query := `DELETE FROM reminders WHERE id = $1`
    result, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        return apperrors.New("INTERNAL_ERROR", "Failed to delete reminder")
    }
    if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
        return apperrors.ReminderNotFound
    }
    return nil
}

// UpdateLastSent updates the last_sent_at timestamp
func (r *ReminderRepository) UpdateLastSent(ctx context.Context, id uuid.UUID, sentAt time.Time) error {
    query := `UPDATE reminders SET last_sent_at = $2, updated_at = $3 WHERE id = $1`
    _, err := r.db.ExecContext(ctx, query, id, sentAt, time.Now())
    if err != nil {
        return apperrors.New("INTERNAL_ERROR", "Failed to update reminder last_sent")
    }
    return nil
}

// GetDueReminders retrieves active reminders that are due
func (r *ReminderRepository) GetDueReminders(ctx context.Context, before time.Time) ([]reminders.Reminder, error) {
    query := `
        SELECT id, user_id, title, description, reminder_time, recurrence, recurrence_value, is_active, last_sent_at, created_at, updated_at
        FROM reminders
        WHERE is_active = true
          AND reminder_time <= $1
          AND (last_sent_at IS NULL OR last_sent_at < reminder_time)
    `
    rows, err := r.db.QueryContext(ctx, query, before)
    if err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to query due reminders")
    }
    defer rows.Close()

    var reminderList []reminders.Reminder
    for rows.Next() {
        var reminder reminders.Reminder
        if err := rows.Scan(
            &reminder.ID,
            &reminder.UserID,
            &reminder.Title,
            &reminder.Description,
            &reminder.ReminderTime,
            &reminder.Recurrence,
            &reminder.RecurrenceValue,
            &reminder.IsActive,
            &reminder.LastSentAt,
            &reminder.CreatedAt,
            &reminder.UpdatedAt,
        ); err != nil {
            return nil, apperrors.New("INTERNAL_ERROR", "Failed to scan reminder row")
        }
        reminderList = append(reminderList, reminder)
    }
    if err := rows.Err(); err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to iterate reminder rows")
    }
    return reminderList, nil
}
