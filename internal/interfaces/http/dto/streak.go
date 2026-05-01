package dto

import (
	"time"

	"github.com/google/uuid"

	"loviary.app/backend/internal/domain/streaks"
)

// StreakResponse represents the streak response DTO.
// Streaks are now per-couple (both users must log for the streak to increment).
type StreakResponse struct {
	StreakID           uuid.UUID           `json:"streak_id"`
	CoupleID           uuid.UUID           `json:"couple_id"`
	ActivityType       string              `json:"activity_type"`
	CurrentStreak      int                 `json:"current_streak"`
	LongestStreak      int                 `json:"longest_streak"`
	LastCompletedDate  *time.Time          `json:"last_completed_date,omitempty"`
	Status             streaks.StreakStatus `json:"status"`
	WeeklyLog          []DayLog            `json:"weekly_log"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
}

// StreakToResponse converts a Streak domain object to response DTO.
func StreakToResponse(s *streaks.Streak) StreakResponse {
	weeklyLog := make([]DayLog, len(s.WeeklyLog))
	for i, d := range s.WeeklyLog {
		weeklyLog[i] = DayLog{Day: d.Day, Completed: d.Completed}
	}

	return StreakResponse{
		StreakID:          s.StreakID,
		CoupleID:          s.CoupleID,
		ActivityType:      s.ActivityType,
		CurrentStreak:     s.CurrentStreak,
		LongestStreak:     s.LongestStreak,
		LastCompletedDate: s.LastCompletedDate,
		Status:            s.GetStreakStatus(),
		WeeklyLog:         weeklyLog,
		CreatedAt:         s.CreatedAt,
		UpdatedAt:         s.UpdatedAt,
	}
}
