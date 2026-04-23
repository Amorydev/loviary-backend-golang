package couples

import (
	"time"

	"github.com/google/uuid"

	"loviary.app/backend/internal/domain/shared"
)

// Couple represents a relationship between two users
type Couple struct {
	CoupleID              uuid.UUID       `db:"couple_id" json:"couple_id"`
	User1ID               uuid.UUID       `db:"user1_id" json:"user1_id"`
	User2ID               *uuid.UUID      `db:"user2_id" json:"user2_id,omitempty"`
	CoupleName            *string         `db:"couple_name" json:"couple_name"`
	RelationshipStartDate *time.Time      `db:"relationship_start_date" json:"relationship_start_date"`
	Status                shared.CoupleStatus `db:"status" json:"status" validate:"required,oneof=pending_invitation active grace_period ended"`
	RelationshipType      shared.RelationshipType `db:"relationship_type" json:"relationship_type" validate:"omitempty,oneof=dating engaged married"`
	InvitationExpiresAt   *time.Time      `db:"invitation_expires_at" json:"invitation_expires_at"`
	BreakupInitiatedBy    *uuid.UUID      `db:"breakup_initiated_by" json:"breakup_initiated_by"`
	BreakupGraceUntil     *time.Time      `db:"breakup_grace_until" json:"breakup_grace_until"`
	CreatedAt             time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time       `db:"updated_at" json:"updated_at"`
}

// IsActive checks if the couple is in active status
func (c *Couple) IsActive() bool {
	return c.Status == shared.CoupleStatusActive
}

// IsPendingInvitation checks if invitation is pending
func (c *Couple) IsPendingInvitation() bool {
	return c.Status == shared.CoupleStatusPendingInvitation
}

// InvitationIsExpired checks if the invitation has expired
func (c *Couple) InvitationIsExpired() bool {
	if c.InvitationExpiresAt == nil {
		return false
	}
	return time.Now().After(*c.InvitationExpiresAt)
}

// GetPartnerID returns the partner's user ID for a given user
func (c *Couple) GetPartnerID(userID uuid.UUID) (uuid.UUID, bool) {
	if c.User1ID == userID && c.User2ID != nil {
		return *c.User2ID, true
	}
	if c.User2ID != nil && *c.User2ID == userID && c.User1ID != uuid.Nil {
		return c.User1ID, true
	}
	return uuid.Nil, false
}
