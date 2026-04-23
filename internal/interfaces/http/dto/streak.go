package dto

import (
    "time"

    "github.com/google/uuid"

    "loviary.app/backend/internal/domain/streaks"
)

// StreakResponse represents streak response DTO
type StreakResponse struct {
    ID              uuid.UUID         `json:"id"`
    UserID          uuid.UUID         `json:"user_id"`
    ActivityType    string            `json:"activity_type"`
    CurrentStreak   int               `json:"current_streak"`
    LongestStreak   int               `json:"longest_streak"`
    LastActiveDate  *time.Time        `json:"last_active_date,omitempty"`
    Status          streaks.StreakStatus `json:"status"`
    CreatedAt       time.Time         `json:"created_at"`
    UpdatedAt       time.Time         `json:"updated_at"`
}

// StreakToResponse converts a Streak to response DTO
func StreakToResponse(streak *streaks.Streak) StreakResponse {
    return StreakResponse{
        ID:             streak.ID,
        UserID:         streak.UserID,
        ActivityType:   streak.ActivityType,
        CurrentStreak:  streak.CurrentStreak,
        LongestStreak:  streak.LongestStreak,
        LastActiveDate: streak.LastActiveDate,
        Status:         streak.GetStreakStatus(),
        CreatedAt:      streak.CreatedAt,
        UpdatedAt:      streak.UpdatedAt,
    }
}
