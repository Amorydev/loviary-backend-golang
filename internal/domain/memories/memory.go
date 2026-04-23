package memories

import (
    "database/sql/driver"
    "encoding/json"
    "errors"
    "time"

    "github.com/google/uuid"
)

// Memory represents a stored memory/moment for a user or couple
type Memory struct {
    ID          uuid.UUID `json:"id" db:"id"`
    UserID      uuid.UUID `json:"user_id" db:"user_id"`
    CoupleID    *uuid.UUID `json:"couple_id,omitempty" db:"couple_id"`
    Title       string    `json:"title" db:"title"`
    Description *string   `json:"description,omitempty" db:"description"`
    MemoryDate  time.Time `json:"memory_date" db:"memory_date"`
    MemoryType  MemoryType `json:"memory_type" db:"memory_type"`
    MediaURLs   []string  `json:"media_urls,omitempty" db:"media_urls"` // JSON array
    Location    *string   `json:"location,omitempty" db:"location"`
    IsPrivate   bool      `json:"is_private" db:"is_private"`
    IsShared    bool      `json:"is_shared" db:"is_shared"` // Shared with partner
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// MemoryType represents the type of memory
type MemoryType string

const (
    MemoryTypeDateNight     MemoryType = "date_night"
    MemoryTypeTrip          MemoryType = "trip"
    MemoryTypeMilestone     MemoryType = "milestone"
    MemoryTypeEveryday      MemoryType = "everyday"
    MemoryTypeCelebration   MemoryType = "celebration"
    MemoryTypeAchievement   MemoryType = "achievement"
)

// Scan implements sql.Scanner for MemoryType
func (mt *MemoryType) Scan(value interface{}) error {
    if value == nil {
        *mt = MemoryTypeEveryday
        return nil
    }
    str, ok := value.(string)
    if !ok {
        return errors.New("invalid type for memory_type")
    }
    *mt = MemoryType(str)
    return nil
}

// Value implements driver.Valuer for MemoryType
func (mt MemoryType) Value() (driver.Value, error) {
    return string(mt), nil
}

// JSONMarshalMemoryType custom marshal for MemoryType
type JSONMarshalMemoryType MemoryType

// MarshalJSON implements json.Marshaler
func (mt MemoryType) MarshalJSON() ([]byte, error) {
    return json.Marshal(string(mt))
}

// UnmarshalJSON implements json.Unmarshaler
func (mt *MemoryType) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }
    *mt = MemoryType(s)
    return nil
}

// IsValid checks if the memory type is valid
func (mt MemoryType) IsValid() bool {
    switch mt {
    case MemoryTypeDateNight, MemoryTypeTrip, MemoryTypeMilestone,
        MemoryTypeEveryday, MemoryTypeCelebration, MemoryTypeAchievement:
        return true
    }
    return false
}
