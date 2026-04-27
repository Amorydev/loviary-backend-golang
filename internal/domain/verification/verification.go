package verification

import (
    "time"

    "github.com/google/uuid"
)

// EmailVerification represents a single email verification record
type EmailVerification struct {
    ID         uuid.UUID  `db:"id" json:"id"`
    UserID     uuid.UUID  `db:"user_id" json:"user_id"`
    Code       string     `db:"code" json:"-"`
    ExpiresAt  time.Time  `db:"expires_at" json:"-"`
    VerifiedAt *time.Time `db:"verified_at" json:"-"`
    CreatedAt  time.Time  `db:"created_at" json:"-"`
}

// IsExpired checks if the verification code has expired
func (v *EmailVerification) IsExpired() bool {
    return time.Now().After(v.ExpiresAt)
}

// IsVerified checks if the email has been verified
func (v *EmailVerification) IsVerified() bool {
    return v.VerifiedAt != nil
}
