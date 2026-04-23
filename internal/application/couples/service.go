package couples

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"loviary.app/backend/internal/domain/couples"
	"loviary.app/backend/internal/domain/shared"
	"loviary.app/backend/pkg/errors"
)

// Repository defines the interface for couple persistence
type Repository interface {
	Create(ctx context.Context, couple *couples.Couple) error
	GetByID(ctx context.Context, id uuid.UUID) (*couples.Couple, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*couples.Couple, error)
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*couples.Couple, error)
	Update(ctx context.Context, couple *couples.Couple) error
	Delete(ctx context.Context, id uuid.UUID) error
	AcceptInvitation(ctx context.Context, coupleID, userID uuid.UUID) error
	InitiateBreakup(ctx context.Context, coupleID, userID uuid.UUID) error
	ConfirmBreakup(ctx context.Context, coupleID uuid.UUID) error
	CancelBreakup(ctx context.Context, coupleID uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]couples.Couple, error)
	Count(ctx context.Context) (int, error)
}

// Service handles couple business logic
type Service struct {
	repo Repository
}

// NewService creates a new couple service
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreateCoupleInput represents input for creating a couple
type CreateCoupleInput struct {
	User1ID    uuid.UUID
	User2ID    uuid.UUID
	CoupleName *string
	RelationshipType shared.RelationshipType
}

// Create creates a new couple relationship
func (s *Service) Create(ctx context.Context, input CreateCoupleInput) (*couples.Couple, error) {
	// Check if either user already has an active couple
	existing1, _ := s.repo.GetActiveByUserID(ctx, input.User1ID)
	if existing1 != nil {
		return nil, errors.New("USER_ALREADY_IN_COUPLE", "User 1 already has an active couple")
	}

	existing2, _ := s.repo.GetActiveByUserID(ctx, input.User2ID)
	if existing2 != nil {
		return nil, errors.New("USER_ALREADY_IN_COUPLE", "User 2 already has an active couple")
	}

	now := time.Now()
	invitationExpires := now.Add(48 * time.Hour)

	couple := &couples.Couple{
		CoupleID:              uuid.New(),
		User1ID:               input.User1ID,
		User2ID:               &input.User2ID,
		CoupleName:            input.CoupleName,
		RelationshipStartDate: &now,
		Status:                shared.CoupleStatusPendingInvitation,
		RelationshipType:      input.RelationshipType,
		InvitationExpiresAt:   &invitationExpires,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if err := s.repo.Create(ctx, couple); err != nil {
		return nil, errors.NewWith("INTERNAL_ERROR", "Failed to create couple", err)
	}

	return couple, nil
}

// GetMyCouple gets the active couple for the current user with partner info
func (s *Service) GetMyCouple(ctx context.Context, userID uuid.UUID) (*couples.Couple, error) {
	couple, err := s.repo.GetActiveByUserID(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NoActiveCouple
		}
		return nil, errors.NewWith("INTERNAL_ERROR", "Failed to get active couple", err)
	}
	return couple, nil
}

// GetByID retrieves a couple by ID
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*couples.Couple, error) {
	couple, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("COUPLE_NOT_FOUND", "Couple not found")
		}
		return nil, errors.NewWith("INTERNAL_ERROR", "Failed to get couple", err)
	}
	return couple, nil
}

// GetActiveByUserID gets the active couple for a user
func (s *Service) GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*couples.Couple, error) {
	couple, err := s.repo.GetActiveByUserID(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NoActiveCouple
		}
		return nil, errors.NewWith("INTERNAL_ERROR", "Failed to get active couple", err)
	}
	return couple, nil
}

// AcceptInvitation accepts a couple invitation
func (s *Service) AcceptInvitation(ctx context.Context, coupleID, userID uuid.UUID) error {
	couple, err := s.repo.GetByID(ctx, coupleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("COUPLE_NOT_FOUND", "Couple not found")
		}
		return errors.NewWith("INTERNAL_ERROR", "Failed to get couple", err)
	}

	// Verify invitation belongs to user
	if couple.User2ID == nil || *couple.User2ID != userID {
		return errors.New("FORBIDDEN", "You are not the invited user")
	}

	// Check if invitation expired
	if couple.InvitationIsExpired() {
		return errors.InvitationExpired
	}

	// Activate the couple
	now := time.Now()
	couple.Status = shared.CoupleStatusActive
	couple.UpdatedAt = now
	couple.InvitationExpiresAt = nil

	return s.repo.Update(ctx, couple)
}

// InitiateBreakup starts the breakup process
func (s *Service) InitiateBreakup(ctx context.Context, coupleID, userID uuid.UUID) error {
	couple, err := s.repo.GetByID(ctx, coupleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("COUPLE_NOT_FOUND", "Couple not found")
		}
		return errors.NewWith("INTERNAL_ERROR", "Failed to get couple", err)
	}

	if !couple.IsActive() {
		return errors.New("INVALID_COUPLE_STATUS", "Couple is not active")
	}

	now := time.Now()
	couple.Status = shared.CoupleStatusGracePeriod
	couple.BreakupInitiatedBy = &userID
	gracePeriod := now.Add(7 * 24 * time.Hour) // 7 days grace period
	couple.BreakupGraceUntil = &gracePeriod
	couple.UpdatedAt = now

	return s.repo.Update(ctx, couple)
}

// ConfirmBreakup confirms the breakup after grace period
func (s *Service) ConfirmBreakup(ctx context.Context, coupleID uuid.UUID) error {
	couple, err := s.repo.GetByID(ctx, coupleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("COUPLE_NOT_FOUND", "Couple not found")
		}
		return errors.NewWith("INTERNAL_ERROR", "Failed to get couple", err)
	}

	if couple.Status != shared.CoupleStatusGracePeriod {
		return errors.New("INVALID_COUPLE_STATUS", "Couple is not in grace period")
	}

	now := time.Now()
	couple.Status = shared.CoupleStatusEnded
	couple.UpdatedAt = now

	return s.repo.Update(ctx, couple)
}

// CancelBreakup reverts couple back to active
func (s *Service) CancelBreakup(ctx context.Context, coupleID uuid.UUID) error {
	couple, err := s.repo.GetByID(ctx, coupleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("COUPLE_NOT_FOUND", "Couple not found")
		}
		return errors.NewWith("INTERNAL_ERROR", "Failed to get couple", err)
	}

	if couple.Status != shared.CoupleStatusGracePeriod {
		return errors.New("INVALID_COUPLE_STATUS", "Couple is not in grace period")
	}

	now := time.Now()
	couple.Status = shared.CoupleStatusActive
	couple.BreakupInitiatedBy = nil
	couple.BreakupGraceUntil = nil
	couple.UpdatedAt = now

	return s.repo.Update(ctx, couple)
}
