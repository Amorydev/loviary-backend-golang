package streaks

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	domainStreaks "loviary.app/backend/internal/domain/streaks"
	apperrors "loviary.app/backend/pkg/errors"
)

// Repository defines the interface for streak persistence.
type Repository interface {
	GetByCoupleAndActivity(ctx context.Context, coupleID uuid.UUID, activityType string) (*domainStreaks.Streak, error)
	GetAllByCouple(ctx context.Context, coupleID uuid.UUID) ([]domainStreaks.Streak, error)
	Upsert(ctx context.Context, streak *domainStreaks.Streak) error
	RecordLog(ctx context.Context, coupleID uuid.UUID, userID uuid.UUID, activityType string, date time.Time) error
	CountLogsForDay(ctx context.Context, coupleID uuid.UUID, activityType string, date time.Time) (int, error)
	GetWeeklyLog(ctx context.Context, coupleID uuid.UUID, activityType string) ([]domainStreaks.DayLog, error)
}

// Service handles streak business logic.
// Streaks are per-couple: both users must log on the same day for current_streak to increment.
type Service struct {
	repo Repository
}

// NewService creates a new streak service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetStreakByCouple retrieves the streak for a couple and activity type, including weekly log.
func (s *Service) GetStreakByCouple(ctx context.Context, coupleID uuid.UUID, activityType string) (*domainStreaks.Streak, error) {
	streak, err := s.repo.GetByCoupleAndActivity(ctx, coupleID, activityType)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			// Return an empty streak if not yet created
			return &domainStreaks.Streak{
				CoupleID:      coupleID,
				ActivityType:  activityType,
				CurrentStreak: 0,
				LongestStreak: 0,
			}, nil
		}
		return nil, apperrors.New("INTERNAL_ERROR", fmt.Sprintf("failed to get streak for couple %s", coupleID))
	}

	weeklyLog, err := s.repo.GetWeeklyLog(ctx, coupleID, activityType)
	if err != nil {
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to get weekly log")
	}
	streak.WeeklyLog = weeklyLog

	return streak, nil
}

// GetAllStreaks retrieves all streaks for a couple (Sprint 1: only mood_log).
func (s *Service) GetAllStreaks(ctx context.Context, coupleID uuid.UUID) ([]domainStreaks.Streak, error) {
	streaks, err := s.repo.GetAllByCouple(ctx, coupleID)
	if err != nil {
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to get streaks")
	}

	for i := range streaks {
		weeklyLog, err := s.repo.GetWeeklyLog(ctx, coupleID, streaks[i].ActivityType)
		if err != nil {
			return nil, apperrors.New("INTERNAL_ERROR", "Failed to get weekly log")
		}
		streaks[i].WeeklyLog = weeklyLog
	}

	return streaks, nil
}

// RecordLog records that a user has completed an activity for today.
// After recording, it checks if both users have logged — if so, increments the couple streak.
func (s *Service) RecordLog(ctx context.Context, coupleID uuid.UUID, userID uuid.UUID, activityType string) error {
	today := time.Now().UTC().Truncate(24 * time.Hour)

	if err := s.repo.RecordLog(ctx, coupleID, userID, activityType, today); err != nil {
		// Duplicate log is not an error (idempotent)
		return nil
	}

	// Check if both users have now logged for today
	count, err := s.repo.CountLogsForDay(ctx, coupleID, activityType, today)
	if err != nil {
		return apperrors.New("INTERNAL_ERROR", "Failed to count daily logs")
	}

	if count >= 2 {
		// Both users logged — increment the couple streak
		if err := s.incrementCoupleStreak(ctx, coupleID, activityType, today); err != nil {
			return err
		}
	}

	return nil
}

// incrementCoupleStreak increments current_streak if today is consecutive with last_completed_date.
func (s *Service) incrementCoupleStreak(ctx context.Context, coupleID uuid.UUID, activityType string, today time.Time) error {
	streak, err := s.repo.GetByCoupleAndActivity(ctx, coupleID, activityType)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.New("INTERNAL_ERROR", "Failed to get streak for increment")
	}

	now := time.Now()

	if streak == nil || errors.Is(err, apperrors.ErrNotFound) {
		// First ever completion
		streak = &domainStreaks.Streak{
			StreakID:           uuid.New(),
			CoupleID:           coupleID,
			ActivityType:       activityType,
			CurrentStreak:      1,
			LongestStreak:      1,
			LastCompletedDate:  &today,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		return s.repo.Upsert(ctx, streak)
	}

	// Already counted today — skip
	if streak.LastCompletedDate != nil {
		lastDate := time.Date(streak.LastCompletedDate.Year(), streak.LastCompletedDate.Month(), streak.LastCompletedDate.Day(), 0, 0, 0, 0, time.UTC)
		todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
		daysDiff := int(todayDate.Sub(lastDate).Hours() / 24)

		if daysDiff == 0 {
			return nil // already counted today
		} else if daysDiff == 1 {
			streak.CurrentStreak++ // consecutive day
		} else {
			streak.CurrentStreak = 1 // streak broken, reset
		}
	} else {
		streak.CurrentStreak = 1
	}

	if streak.CurrentStreak > streak.LongestStreak {
		streak.LongestStreak = streak.CurrentStreak
	}

	streak.LastCompletedDate = &today
	streak.UpdatedAt = now

	return s.repo.Upsert(ctx, streak)
}
