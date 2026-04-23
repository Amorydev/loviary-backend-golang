package home

import (
    "context"
    "errors"
    "time"

    "github.com/google/uuid"

    "loviary.app/backend/internal/application/couples"
    "loviary.app/backend/internal/application/moods"
    "loviary.app/backend/internal/application/streaks"
    domainCouples "loviary.app/backend/internal/domain/couples"
    domainMoods "loviary.app/backend/internal/domain/moods"
    domainStreaks "loviary.app/backend/internal/domain/streaks"
    apperrors "loviary.app/backend/pkg/errors"
)

// Service aggregates data for the home dashboard
type Service struct {
    coupleService *couples.Service
    moodService   *moods.Service
    streakService *streaks.Service
}

// NewService creates a new home service
func NewService(
    coupleService *couples.Service,
    moodService *moods.Service,
    streakService *streaks.Service,
) *Service {
    return &Service{
        coupleService: coupleService,
        moodService:   moodService,
        streakService: streakService,
    }
}

// DashboardData represents aggregated dashboard data
type DashboardData struct {
    UserID           uuid.UUID             `json:"user_id"`
    Couple           *domainCouples.Couple `json:"couple,omitempty"`
    TodaysMood       *domainMoods.Mood     `json:"todays_mood,omitempty"`
    MoodHistory      []domainMoods.Mood    `json:"mood_history,omitempty"`
    Streaks          []domainStreaks.Streak `json:"streaks,omitempty"`
    RecentMemories   interface{}           `json:"recent_memories,omitempty"` // Will be added when memories service is ready
    UpcomingReminders interface{}          `json:"upcoming_reminders,omitempty"` // Will be added when reminders service is ready
    LastUpdated      time.Time             `json:"last_updated"`
}

// GetDashboard retrieves aggregated data for the user's home dashboard
func (s *Service) GetDashboard(ctx context.Context, userID uuid.UUID) (*DashboardData, error) {
    data := &DashboardData{
        UserID: userID,
    }

    // Get user's active couple
    couple, err := s.coupleService.GetActiveByUserID(ctx, userID)
    if err != nil {
        // No active couple is not an error, just nil
        if !errors.Is(err, apperrors.NoActiveCouple) && !errors.Is(err, apperrors.ErrNotFound) {
            return nil, apperrors.New("INTERNAL_ERROR", "Failed to get couple")
        }
    } else {
        data.Couple = couple
    }

    // Get today's mood
    todaysMood, err := s.moodService.GetTodaysMood(ctx, userID)
    if err != nil {
        // Not an error if no mood logged today
        if !errors.Is(err, apperrors.ErrNotFound) {
            return nil, apperrors.New("INTERNAL_ERROR", "Failed to get today's mood")
        }
    } else {
        data.TodaysMood = todaysMood
    }

    // Get mood history (last 7 days)
    endDate := time.Now()
    startDate := endDate.Add(-7 * 24 * time.Hour)
    moodHistory, err := s.moodService.GetMoodsByUser(ctx, userID, startDate, endDate)
    if err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get mood history")
    }
    data.MoodHistory = moodHistory

    // Get all streaks
    streaks, err := s.streakService.GetAllStreaks(ctx, userID)
    if err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get streaks")
    }
    data.Streaks = streaks

    // TODO: Add recent memories and upcoming reminders when those services are integrated

    data.LastUpdated = time.Now()

    return data, nil
}

// HomeSummary represents a summary for the home endpoint
type HomeSummary struct {
    CoupleID         *uuid.UUID          `json:"couple_id,omitempty"`
    PartnerName      *string             `json:"partner_name,omitempty"`
    RelationshipType *string             `json:"relationship_type,omitempty"`
    TodaysMoodType  *string             `json:"todays_mood_type,omitempty"`
    TodaysMoodIntensity *int            `json:"todays_mood_intensity,omitempty"`
    ActiveStreaks   int                 `json:"active_streaks"`
    LongestStreak   int                 `json:"longest_streak"`
    RecentActivity  []string            `json:"recent_activity,omitempty"`
}

// GetSummary retrieves a simplified summary for the home endpoint
func (s *Service) GetSummary(ctx context.Context, userID uuid.UUID) (*HomeSummary, error) {
    summary := &HomeSummary{}

    // Get couple info
    couple, err := s.coupleService.GetActiveByUserID(ctx, userID)
    if err != nil {
        if !errors.Is(err, apperrors.NoActiveCouple) && !errors.Is(err, apperrors.ErrNotFound) {
            return nil, apperrors.New("INTERNAL_ERROR", "Failed to get couple")
        }
    } else {
        summary.CoupleID = &couple.CoupleID
        summary.RelationshipType = (*string)(&couple.RelationshipType)
        // TODO: Get partner name when user service is available
    }

    // Get today's mood
    todaysMood, err := s.moodService.GetTodaysMood(ctx, userID)
    if err != nil {
        if !errors.Is(err, apperrors.ErrNotFound) {
            return nil, apperrors.New("INTERNAL_ERROR", "Failed to get today's mood")
        }
    } else {
        summary.TodaysMoodType = (*string)(&todaysMood.MoodType)
        summary.TodaysMoodIntensity = &todaysMood.Intensity
    }

    // Get streaks
    allStreaks, err := s.streakService.GetAllStreaks(ctx, userID)
    if err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get streaks")
    }

    activeCount := 0
    longest := 0
    for _, streak := range allStreaks {
        if streak.CurrentStreak > 0 {
            activeCount++
        }
        if streak.LongestStreak > longest {
            longest = streak.LongestStreak
        }
    }
    summary.ActiveStreaks = activeCount
    summary.LongestStreak = longest

    return summary, nil
}
