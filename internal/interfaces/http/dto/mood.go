package dto

import (
    "time"

    "github.com/google/uuid"

    "loviary.app/backend/internal/domain/moods"
)

// MoodResponse represents mood response DTO
type MoodResponse struct {
    ID          uuid.UUID            `json:"id"`
    UserID      uuid.UUID            `json:"user_id"`
    Date        time.Time            `json:"date"`
    MoodType    moods.MoodType       `json:"mood_type"`
    Intensity   int                  `json:"intensity"`
    MoodEmoji   *string              `json:"mood_emoji,omitempty"`
    Note        *string              `json:"note,omitempty"`
    IsShared    bool                 `json:"is_shared"`
    CreatedAt   time.Time            `json:"created_at"`
    UpdatedAt   time.Time            `json:"updated_at"`
}

// MoodToResponse converts a Mood to response DTO
func MoodToResponse(mood *moods.Mood) MoodResponse {
    return MoodResponse{
        ID:         mood.ID,
        UserID:     mood.UserID,
        Date:       mood.Date,
        MoodType:   mood.MoodType,
        Intensity:  mood.Intensity,
        MoodEmoji:  mood.MoodEmoji,
        Note:       mood.Note,
        IsShared:   mood.IsShared,
        CreatedAt:  mood.CreatedAt,
        UpdatedAt:  mood.UpdatedAt,
    }
}
