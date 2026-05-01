package chapters

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	domainChapters "loviary.app/backend/internal/domain/chapters"
	apperrors "loviary.app/backend/pkg/errors"
)

// Repository defines the interface for chapter persistence.
type Repository interface {
	GetDefinitionByDay(ctx context.Context, daysTogether int) (*domainChapters.ChapterDefinition, error)
	GetAllDefinitions(ctx context.Context) ([]domainChapters.ChapterDefinition, error)
	GetCoupleChapter(ctx context.Context, coupleID uuid.UUID, chapterNumber int) (*domainChapters.CoupleChapter, error)
	UnlockChapter(ctx context.Context, coupleID uuid.UUID, chapterNumber int, unlockedBy uuid.UUID) error
}

// Service handles chapter business logic.
type Service struct {
	repo Repository
}

// NewService creates a new chapter service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetCurrentChapter returns the current chapter view for a couple based on days together.
func (s *Service) GetCurrentChapter(ctx context.Context, coupleID uuid.UUID, daysTogether int) (*domainChapters.CurrentChapterView, error) {
	def, err := s.repo.GetDefinitionByDay(ctx, daysTogether)
	if err != nil {
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to get chapter definition")
	}

	view := &domainChapters.CurrentChapterView{
		Definition:       *def,
		DaysTogether:     daysTogether,
		MilestonePercent: def.CalculateMilestonePercent(daysTogether),
	}

	// Check if this chapter is already unlocked for this couple
	coupleChapter, err := s.repo.GetCoupleChapter(ctx, coupleID, def.ChapterNumber)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to get couple chapter state")
	}
	if coupleChapter != nil {
		view.IsUnlocked = coupleChapter.IsUnlocked
	}

	return view, nil
}

// CheckAndUnlockChapter auto-unlocks a chapter if the couple has reached its day_start.
// Should be called after days_together is calculated.
func (s *Service) CheckAndUnlockChapter(ctx context.Context, coupleID uuid.UUID, daysTogether int, triggeredBy uuid.UUID) error {
	def, err := s.repo.GetDefinitionByDay(ctx, daysTogether)
	if err != nil {
		return apperrors.New("INTERNAL_ERROR", "Failed to get chapter definition")
	}

	existing, err := s.repo.GetCoupleChapter(ctx, coupleID, def.ChapterNumber)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return apperrors.New("INTERNAL_ERROR", "Failed to check chapter unlock state")
	}

	// Already unlocked — nothing to do
	if existing != nil && existing.IsUnlocked {
		return nil
	}

	// Unlock if couple has reached or passed the chapter start
	if daysTogether >= def.DayStart {
		if err := s.repo.UnlockChapter(ctx, coupleID, def.ChapterNumber, triggeredBy); err != nil {
			return apperrors.New("INTERNAL_ERROR", "Failed to unlock chapter")
		}
	}

	return nil
}

// GetAllDefinitions returns all chapter definitions ordered by chapter_number.
func (s *Service) GetAllDefinitions(ctx context.Context) ([]domainChapters.ChapterDefinition, error) {
	defs, err := s.repo.GetAllDefinitions(ctx)
	if err != nil {
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to get chapter definitions")
	}
	return defs, nil
}

// CalculateDaysTogether returns the number of days since the relationship start date.
func CalculateDaysTogether(startDate time.Time) int {
	now := time.Now()
	start := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	days := int(today.Sub(start).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}
