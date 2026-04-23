package shared

import (
	"time"

	"github.com/google/uuid"
)

type ID = uuid.UUID

// NewID creates a new UUID
func NewID() ID {
	return uuid.New()
}

// MustParseID parses a UUID or panics
func MustParseID(s string) ID {
	id, err := uuid.Parse(s)
	if err != nil {
		panic(err)
	}
	return id
}

// Timestamp represents a time with UTC normalization
type Timestamp time.Time

func (t Timestamp) Time() time.Time {
	return time.Time(t).UTC()
}

func NewTimestamp() Timestamp {
	return Timestamp(time.Now().UTC())
}

// Gender represents user gender
type Gender string

const (
	GenderMale          Gender = "male"
	GenderFemale        Gender = "female"
	GenderOther         Gender = "other"
	GenderPreferNot     Gender = "prefer_not"
)

// UserStatus represents user account status
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
)

// CoupleStatus represents relationship status
type CoupleStatus string

const (
	CoupleStatusPendingInvitation CoupleStatus = "pending_invitation"
	CoupleStatusActive            CoupleStatus = "active"
	CoupleStatusGracePeriod       CoupleStatus = "grace_period"
	CoupleStatusEnded             CoupleStatus = "ended"
)

// RelationshipType represents the type of relationship
type RelationshipType string

const (
	RelationshipDating   RelationshipType = "dating"
	RelationshipEngaged  RelationshipType = "engaged"
	RelationshipMarried  RelationshipType = "married"
)

// SubscriptionPlan represents subscription tier
type SubscriptionPlan string

const (
	PlanFree         SubscriptionPlan = "free"
	PlanPremium      SubscriptionPlan = "premium"
	PlanPremiumPlus  SubscriptionPlan = "premium_plus"
	PlanLifetime     SubscriptionPlan = "lifetime"
	PlanSoloArchive  SubscriptionPlan = "solo_archive"
)

// MoodType represents the mood logged
type MoodType string

const (
	MoodHappy    MoodType = "happy"
	MoodSad      MoodType = "sad"
	MoodAngry    MoodType = "angry"
	MoodNeutral  MoodType = "neutral"
	MoodExcited  MoodType = "excited"
)

// NotificationType represents notification category
type NotificationType string

const (
	NotificationMoodSync   NotificationType = "mood_sync"
	NotificationStreak    NotificationType = "streak"
	NotificationReminder  NotificationType = "reminder"
	NotificationSpark     NotificationType = "spark"
	NotificationCapsule   NotificationType = "capsule"
	NotificationChapter   NotificationType = "chapter"
	NotificationSOS       NotificationType = "sos"
	NotificationPairInvite NotificationType = "pair_invite"
)
