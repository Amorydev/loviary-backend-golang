package verification

import (
    "context"
    "errors"
    "fmt"
    "math/rand"
    "time"

    "github.com/google/uuid"

    "loviary.app/backend/internal/domain/verification"
    "loviary.app/backend/internal/infrastructure/email"
    "loviary.app/backend/pkg/logger"
)

// Repository interface for email verification persistence
type Repository interface {
    Create(ctx context.Context, v *verification.EmailVerification) error
    GetByUserID(ctx context.Context, userID uuid.UUID) (*verification.EmailVerification, error)
    GetByCode(ctx context.Context, code string) (*verification.EmailVerification, error)
    Verify(ctx context.Context, id uuid.UUID) error
    Delete(ctx context.Context, userID uuid.UUID) error
}

// Domain errors
var (
    ErrNotFound        = errors.New("NOT_FOUND")
    ErrExpired         = errors.New("EXPIRED")
    ErrAlreadyVerified = errors.New("ALREADY_VERIFIED")
    ErrResendTooFast   = errors.New("RESEND_TOO_FAST")
)

// Service handles email verification business logic
type Service struct {
    repo           Repository
    emailSender    email.Sender
    log            *logger.Logger
    codeTTL        time.Duration
    resendWindow   time.Duration
}

// NewService creates a new verification service
func NewService(
    repo Repository,
    emailSender email.Sender,
    log *logger.Logger,
    codeTTL, resendWindow time.Duration,
) *Service {
    return &Service{
        repo:          repo,
        emailSender:   emailSender,
        log:           log,
        codeTTL:       codeTTL,
        resendWindow:  resendWindow,
    }
}

// generateCode creates a random 6-digit code
func (s *Service) generateCode() string {
    // Generate random number between 100000 and 999999
    return fmt.Sprintf("%06d", rand.Intn(900000)+100000)
}

// CreateVerification generates a new verification code and sends email
func (s *Service) CreateVerification(ctx context.Context, userID uuid.UUID, userEmail string) error {
    // Check if there's an existing unverified, non-expired verification
    existing, _ := s.repo.GetByUserID(ctx, userID)
    if existing != nil && !existing.IsVerified() && !existing.IsExpired() {
        return ErrResendTooFast // or could be a different error like "pending exists"
    }

    // Generate new code
    code := s.generateCode()
    now := time.Now()
    v := &verification.EmailVerification{
        ID:         uuid.New(),
        UserID:     userID,
        Code:       code,
        ExpiresAt:  now.Add(s.codeTTL),
        CreatedAt:  now,
    }

    // Save to database
    if err := s.repo.Create(ctx, v); err != nil {
        return err
    }

    // Send email
    s.log.Info("Sending verification email", map[string]interface{}{
        "user_id": userID,
        "email":   userEmail,
        "code":    code,
    })
    if err := s.emailSender.SendVerificationEmail(userEmail, code); err != nil {
        s.log.Error("Failed to send verification email", err, map[string]interface{}{
            "user_id": userID,
            "email":   userEmail,
        })
        // Don't rollback DB, but log the error
        // Could optionally delete the verification record here
    } else {
        s.log.Info("Verification email sent successfully", map[string]interface{}{
            "user_id": userID,
            "email":   userEmail,
        })
    }
    return nil
}

// VerifyCode validates the verification code and marks it as verified
func (s *Service) VerifyCode(ctx context.Context, code string) (*verification.EmailVerification, error) {
    v, err := s.repo.GetByCode(ctx, code)
    if err != nil {
        return nil, err
    }

    if v.IsExpired() {
        return nil, ErrExpired
    }

    if v.IsVerified() {
        return nil, ErrAlreadyVerified
    }

    // Mark as verified
    if err := s.repo.Verify(ctx, v.ID); err != nil {
        return nil, fmt.Errorf("failed to verify code: %w", err)
    }

    s.log.Info("Email verified", map[string]interface{}{
        "user_id": v.UserID,
    })

    return v, nil
}

// CanResend checks if a new verification code can be sent
// Returns: (canResend, nextAvailableTime, error)
func (s *Service) CanResend(ctx context.Context, userID uuid.UUID) (bool, time.Time, error) {
    existing, _ := s.repo.GetByUserID(ctx, userID)
    if existing == nil {
        return false, time.Time{}, ErrNotFound
    }

    if existing.IsVerified() {
        return false, time.Time{}, ErrAlreadyVerified
    }

    if existing.IsExpired() {
        return false, time.Time{}, ErrExpired
    }

    // Check resend window: must wait at least resendWindow from created_at
    timeSinceCreated := time.Since(existing.CreatedAt)
    if timeSinceCreated < s.resendWindow {
        next := existing.CreatedAt.Add(s.resendWindow)
        return false, next, nil
    }

    return true, time.Time{}, nil
}
