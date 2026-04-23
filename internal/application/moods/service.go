package moods

import (
    "context"
    "errors"
    "time"

    "github.com/google/uuid"

    "loviary.app/backend/internal/domain/moods"
    apperrors "loviary.app/backend/pkg/errors"
)

// Repository defines the interface for mood persistence
type Repository interface {
    Create(ctx context.Context, mood *moods.Mood) error
    GetByID(ctx context.Context, id uuid.UUID) (*moods.Mood, error)
    GetByUserIDAndDate(ctx context.Context, userID uuid.UUID, date time.Time) (*moods.Mood, error)
    GetByUserID(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]moods.Mood, error)
    Update(ctx context.Context, mood *moods.Mood) error
    Delete(ctx context.Context, id uuid.UUID) error
    GetSharedMoods(ctx context.Context, userID1, userID2 uuid.UUID, startDate, endDate time.Time) ([]moods.Mood, error)
}

// Service handles mood business logic
type Service struct {
    repo Repository
}

// NewService creates a new mood service
func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}

// CreateMoodInput represents input for creating a mood
type CreateMoodInput struct {
    UserID    uuid.UUID
    Date      time.Time
    MoodType  moods.MoodType
    Intensity int
    Note      *string
    IsShared  bool
}

// Create creates a new mood entry
func (s *Service) Create(ctx context.Context, input CreateMoodInput) (*moods.Mood, error) {
    // Validate mood type
    if !input.MoodType.IsValid() {
        return nil, apperrors.New("INVALID_MOOD_TYPE", "Invalid mood type")
    }

    // Validate intensity (1-10)
    if input.Intensity < 1 || input.Intensity > 10 {
        return nil, apperrors.New("INVALID_INTENSITY", "Intensity must be between 1 and 10")
    }

    // Normalize date to start of day
    date := input.Date.Truncate(24 * time.Hour)

    // Check if mood already exists for this user on this date
    existing, _ := s.repo.GetByUserIDAndDate(ctx, input.UserID, date)
    if existing != nil {
        return nil, apperrors.New("MOOD_ALREADY_EXISTS", "Mood already logged for this date")
    }

    now := time.Now()
    mood := &moods.Mood{
        ID:          uuid.New(),
        UserID:      input.UserID,
        Date:        date,
        MoodType:    input.MoodType,
        Intensity:   input.Intensity,
        Note:        input.Note,
        IsShared:    input.IsShared,
        CreatedAt:   now,
        UpdatedAt:   now,
    }

    if err := s.repo.Create(ctx, mood); err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to create mood")
    }

    return mood, nil
}

// UpdateMoodInput represents input for updating a mood
type UpdateMoodInput struct {
    MoodType  *moods.MoodType
    Intensity *int
    Note      *string
    IsShared  *bool
}

// Update updates an existing mood
func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateMoodInput) (*moods.Mood, error) {
    mood, err := s.repo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, apperrors.New("NOT_FOUND", "")) {
            return nil, apperrors.New("MOOD_NOT_FOUND", "Mood not found")
        }
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get mood")
    }

    // Update fields if provided
    if input.MoodType != nil {
        if !(*input.MoodType).IsValid() {
            return nil, apperrors.New("INVALID_MOOD_TYPE", "Invalid mood type")
        }
        mood.MoodType = *input.MoodType
    }
    if input.Intensity != nil {
        intensity := *input.Intensity
        if intensity < 1 || intensity > 10 {
            return nil, apperrors.New("INVALID_INTENSITY", "Intensity must be between 1 and 10")
        }
        mood.Intensity = intensity
    }
    if input.Note != nil {
        mood.Note = input.Note
    }
    if input.IsShared != nil {
        mood.IsShared = *input.IsShared
    }

    mood.UpdatedAt = time.Now()

    if err := s.repo.Update(ctx, mood); err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to update mood")
    }

    return mood, nil
}

// GetMood retrieves a mood by ID
func (s *Service) GetMood(ctx context.Context, id uuid.UUID) (*moods.Mood, error) {
    mood, err := s.repo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, apperrors.New("NOT_FOUND", "")) {
            return nil, apperrors.New("MOOD_NOT_FOUND", "Mood not found")
        }
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get mood")
    }
    return mood, nil
}

// GetMoodsByUser retrieves moods for a user within a date range
func (s *Service) GetMoodsByUser(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]moods.Mood, error) {
    // Default to last 30 days if not specified
    if startDate.IsZero() || endDate.IsZero() {
        endDate = time.Now()
        startDate = endDate.Add(-30 * 24 * time.Hour)
    }

    moodList, err := s.repo.GetByUserID(ctx, userID, startDate, endDate)
    if err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get moods")
    }
    return moodList, nil
}

// GetTodaysMood retrieves today's mood for a user
func (s *Service) GetTodaysMood(ctx context.Context, userID uuid.UUID) (*moods.Mood, error) {
    today := time.Now().Truncate(24 * time.Hour)
    mood, err := s.repo.GetByUserIDAndDate(ctx, userID, today)
    if err != nil {
        if errors.Is(err, apperrors.MoodNotFound) {
            return nil, nil // No mood logged today
        }
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get today's mood")
    }
    return mood, nil
}

// DeleteMood removes a mood
func (s *Service) DeleteMood(ctx context.Context, id uuid.UUID) error {
    // First verify the mood exists
    _, err := s.repo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, apperrors.MoodNotFound) {
            return apperrors.New("MOOD_NOT_FOUND", "Mood not found")
        }
        return apperrors.New("INTERNAL_ERROR", "Failed to get mood")
    }

    if err := s.repo.Delete(ctx, id); err != nil {
        return apperrors.New("INTERNAL_ERROR", "Failed to delete mood")
    }
    return nil
}

// GetSharedMoods retrieves shared moods for a couple
func (s *Service) GetSharedMoods(ctx context.Context, userID1, userID2 uuid.UUID, startDate, endDate time.Time) ([]moods.Mood, error) {
    if startDate.IsZero() || endDate.IsZero() {
        endDate = time.Now()
        startDate = endDate.Add(-30 * 24 * time.Hour)
    }

    moodList, err := s.repo.GetSharedMoods(ctx, userID1, userID2, startDate, endDate)
    if err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get shared moods")
    }
    return moodList, nil
}
