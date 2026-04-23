package streaks

import (
    "context"
    "errors"

    "github.com/google/uuid"

    "loviary.app/backend/internal/domain/streaks"
    apperrors "loviary.app/backend/pkg/errors"
)

// Repository defines the interface for streak persistence
type Repository interface {
    Create(ctx context.Context, streak *streaks.Streak) error
    GetByID(ctx context.Context, id uuid.UUID) (*streaks.Streak, error)
    GetByUserIDAndActivity(ctx context.Context, userID uuid.UUID, activityType string) (*streaks.Streak, error)
    Update(ctx context.Context, streak *streaks.Streak) error
    IncrementStreak(ctx context.Context, userID uuid.UUID, activityType string) (*streaks.Streak, error)
    GetByUserID(ctx context.Context, userID uuid.UUID) ([]streaks.Streak, error)
}

// Service handles streak business logic
type Service struct {
    repo Repository
}

// NewService creates a new streak service
func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}

// RecordActivity records an activity and updates the streak
func (s *Service) RecordActivity(ctx context.Context, userID uuid.UUID, activityType string) (*streaks.Streak, error) {
    streak, err := s.repo.IncrementStreak(ctx, userID, activityType)
    if err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to record activity")
    }
    return streak, nil
}

// GetStreak retrieves a specific streak for a user
func (s *Service) GetStreak(ctx context.Context, userID uuid.UUID, activityType string) (*streaks.Streak, error) {
    streak, err := s.repo.GetByUserIDAndActivity(ctx, userID, activityType)
    if err != nil {
        if errors.Is(err, apperrors.StreakNotFound) {
            return nil, nil // No streak yet
        }
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get streak")
    }
    return streak, nil
}

// GetAllStreaks retrieves all streaks for a user
func (s *Service) GetAllStreaks(ctx context.Context, userID uuid.UUID) ([]streaks.Streak, error) {
    streaks, err := s.repo.GetByUserID(ctx, userID)
    if err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get streaks")
    }
    return streaks, nil
}

// GetStreakStatus returns the status of a streak
func (s *Service) GetStreakStatus(ctx context.Context, userID uuid.UUID, activityType string) (streaks.StreakStatus, error) {
    streak, err := s.repo.GetByUserIDAndActivity(ctx, userID, activityType)
    if err != nil {
        if errors.Is(err, apperrors.StreakNotFound) {
            return streaks.StreakStatusNone, nil
        }
        return "", apperrors.New("INTERNAL_ERROR", "Failed to get streak")
    }
    return streak.GetStreakStatus(), nil
}
