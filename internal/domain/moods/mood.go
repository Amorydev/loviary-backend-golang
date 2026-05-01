package moods

import (
    "database/sql/driver"
    "encoding/json"
    "errors"
    "time"

    "github.com/google/uuid"
)

// Mood represents a user's mood entry
type Mood struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Date      time.Time `json:"date" db:"date"`
	MoodType  MoodType  `json:"mood_type" db:"mood_type"`
	Intensity int       `json:"intensity" db:"intensity"` // 1-10 scale
	MoodEmoji *string   `json:"mood_emoji,omitempty" db:"mood_emoji"`
	Note      *string   `json:"note,omitempty" db:"note"`
	IsShared  bool      `json:"is_shared" db:"is_shared"` // Shared with partner
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// MoodType represents the type of mood
type MoodType string

const (
    MoodTypeHappy     MoodType = "happy"
    MoodTypeSad       MoodType = "sad"
    MoodTypeAngry     MoodType = "angry"
    MoodTypeAnxious   MoodType = "anxious"
    MoodTypeCalm      MoodType = "calm"
    MoodTypeExcited   MoodType = "excited"
    MoodTypeTired     MoodType = "tired"
    MoodTypeLove      MoodType = "love"
    MoodTypeStressed  MoodType = "stressed"
    MoodTypeGrateful  MoodType = "grateful"
)

// Scan implements sql.Scanner for MoodType
func (mt *MoodType) Scan(value interface{}) error {
    if value == nil {
        *mt = MoodTypeHappy
        return nil
    }
    str, ok := value.(string)
    if !ok {
        return errors.New("invalid type for mood_type")
    }
    *mt = MoodType(str)
    return nil
}

// Value implements driver.Valuer for MoodType
func (mt MoodType) Value() (driver.Value, error) {
    return string(mt), nil
}

// JSONMarshalMoodType custom marshal for MoodType
type JSONMarshalMoodType MoodType

// MarshalJSON implements json.Marshaler
func (mt MoodType) MarshalJSON() ([]byte, error) {
    return json.Marshal(string(mt))
}

// UnmarshalJSON implements json.Unmarshaler
func (mt *MoodType) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }
    *mt = MoodType(s)
    return nil
}

// IsValid checks if the mood type is valid
func (mt MoodType) IsValid() bool {
    switch mt {
    case MoodTypeHappy, MoodTypeSad, MoodTypeAngry, MoodTypeAnxious,
        MoodTypeCalm, MoodTypeExcited, MoodTypeTired, MoodTypeLove,
        MoodTypeStressed, MoodTypeGrateful:
        return true
    }
    return false
}
