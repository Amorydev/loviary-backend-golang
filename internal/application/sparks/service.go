package sparks

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	domainSparks "loviary.app/backend/internal/domain/sparks"
	apperrors "loviary.app/backend/pkg/errors"
)

// Repository defines the interface for spark persistence.
type Repository interface {
	GetTodayAssignment(ctx context.Context, coupleID uuid.UUID, date time.Time) (*domainSparks.SparkAssignment, error)
	GetSparkByID(ctx context.Context, sparkID uuid.UUID) (*domainSparks.DailySpark, error)
	GetUserResponse(ctx context.Context, sparkID uuid.UUID, userID uuid.UUID) (*domainSparks.SparkResponse, error)
	GetRandomSpark(ctx context.Context) (*domainSparks.DailySpark, error)
	AssignSpark(ctx context.Context, assignment *domainSparks.SparkAssignment) error
}

// Service handles daily spark business logic.
type Service struct {
	repo Repository
}

// NewService creates a new spark service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetTodaySpark returns the daily spark for a couple and whether the user has answered it.
// If no spark is assigned for today, it assigns one automatically.
func (s *Service) GetTodaySpark(ctx context.Context, coupleID uuid.UUID, userID uuid.UUID) (*domainSparks.DailySpark, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)

	assignment, err := s.repo.GetTodayAssignment(ctx, coupleID, today)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			// Auto-assign if none exists
			if assignErr := s.AssignDailySpark(ctx, coupleID); assignErr != nil {
				return nil, assignErr
			}
			assignment, err = s.repo.GetTodayAssignment(ctx, coupleID, today)
			if err != nil {
				return nil, apperrors.New("INTERNAL_ERROR", "Failed to get today's spark after assignment")
			}
		} else {
			return nil, apperrors.New("INTERNAL_ERROR", "Failed to get today's spark assignment")
		}
	}

	spark, err := s.repo.GetSparkByID(ctx, assignment.SparkID)
	if err != nil {
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to get spark details")
	}

	// Check if user has already answered
	response, err := s.repo.GetUserResponse(ctx, assignment.SparkID, userID)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to check spark response")
	}
	spark.IsAnswered = response != nil

	return spark, nil
}

// AssignDailySpark picks a random spark and assigns it to the couple for today.
func (s *Service) AssignDailySpark(ctx context.Context, coupleID uuid.UUID) error {
	spark, err := s.repo.GetRandomSpark(ctx)
	if err != nil {
		return apperrors.New("INTERNAL_ERROR", "Failed to get random spark")
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	assignment := &domainSparks.SparkAssignment{
		AssignmentID: uuid.New(),
		SparkID:      spark.SparkID,
		CoupleID:     coupleID,
		AssignedDate: today,
		ExpiresAt:    today.Add(24 * time.Hour),
	}

	if err := s.repo.AssignSpark(ctx, assignment); err != nil {
		return apperrors.New("INTERNAL_ERROR", "Failed to assign spark")
	}

	return nil
}
