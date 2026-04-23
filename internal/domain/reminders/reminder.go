package reminders

import (
    "database/sql/driver"
    "encoding/json"
    "errors"
    "time"

    "github.com/google/uuid"
)

// Reminder represents a reminder for an activity
type Reminder struct {
    ID          uuid.UUID `json:"id" db:"id"`
    UserID      uuid.UUID `json:"user_id" db:"user_id"`
    Title       string    `json:"title" db:"title"`
    Description *string   `json:"description,omitempty" db:"description"`
    ReminderTime time.Time `json:"reminder_time" db:"reminder_time"`
    Recurrence  RecurrenceType `json:"recurrence" db:"recurrence"`
    RecurrenceValue *int `json:"recurrence_value,omitempty" db:"recurrence_value"` // e.g., every X days/weeks
    IsActive    bool      `json:"is_active" db:"is_active"`
    LastSentAt  *time.Time `json:"last_sent_at,omitempty" db:"last_sent_at"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// RecurrenceType represents how often a reminder recurs
type RecurrenceType string

const (
    RecurrenceNone     RecurrenceType = "none"
    RecurrenceDaily    RecurrenceType = "daily"
    RecurrenceWeekly   RecurrenceType = "weekly"
    RecurrenceMonthly  RecurrenceType = "monthly"
    RecurrenceCustom   RecurrenceType = "custom"
)

// Scan implements sql.Scanner for RecurrenceType
func (rt *RecurrenceType) Scan(value interface{}) error {
    if value == nil {
        *rt = RecurrenceNone
        return nil
    }
    str, ok := value.(string)
    if !ok {
        return errors.New("invalid type for recurrence_type")
    }
    *rt = RecurrenceType(str)
    return nil
}

// Value implements driver.Valuer for RecurrenceType
func (rt RecurrenceType) Value() (driver.Value, error) {
    return string(rt), nil
}

// JSONMarshalRecurrenceType custom marshal for RecurrenceType
type JSONMarshalRecurrenceType RecurrenceType

// MarshalJSON implements json.Marshaler
func (rt RecurrenceType) MarshalJSON() ([]byte, error) {
    return json.Marshal(string(rt))
}

// UnmarshalJSON implements json.Unmarshaler
func (rt *RecurrenceType) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }
    *rt = RecurrenceType(s)
    return nil
}

// IsValid checks if the recurrence type is valid
func (rt RecurrenceType) IsValid() bool {
    switch rt {
    case RecurrenceNone, RecurrenceDaily, RecurrenceWeekly,
        RecurrenceMonthly, RecurrenceCustom:
        return true
    }
    return false
}
