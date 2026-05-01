package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	domainAuth "loviary.app/backend/internal/domain/auth"
	"loviary.app/backend/internal/domain/users"
	"loviary.app/backend/internal/application/verification"
	"loviary.app/backend/internal/infrastructure/jwt"
	postgres "loviary.app/backend/internal/infrastructure/persistence/postgres"
	apperrors "loviary.app/backend/pkg/errors"
	"loviary.app/backend/pkg/logger"
)

// Service handles authentication business logic
type Service struct {
	userRepo            *postgres.UserRepository
	tokenRepo           *postgres.RefreshTokenRepository
	coupleRepo          *postgres.CoupleRepository
	jwtManager          *jwt.Manager
	verificationService *verification.Service
	accessTokenTTL      time.Duration
	refreshTokenTTL     time.Duration
	log                 *logger.Logger
}

// NewService creates a new auth service
func NewService(
	userRepo *postgres.UserRepository,
	tokenRepo *postgres.RefreshTokenRepository,
	coupleRepo *postgres.CoupleRepository,
	jwtManager *jwt.Manager,
	verificationService *verification.Service,
	log *logger.Logger,
	accessTokenTTL, refreshTokenTTL time.Duration,
) *Service {
	return &Service{
		userRepo:            userRepo,
		tokenRepo:           tokenRepo,
		coupleRepo:          coupleRepo,
		jwtManager:          jwtManager,
		verificationService: verificationService,
		accessTokenTTL:      accessTokenTTL,
		refreshTokenTTL:     refreshTokenTTL,
		log:                 log,
	}
}

// LoginInput represents login credentials
type LoginInput struct {
    Email      string
    Password   string
    DeviceInfo string
}

// LoginResult contains tokens after successful login
type LoginResult struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int64       `json:"expires_in"`
	User         *users.User `json:"user"`
	HasCouple    bool        `json:"has_couple"`
}

// Auth errors
var (
    ErrInvalidCredentials = apperrors.New("INVALID_CREDENTIALS", "Invalid email or password")
    ErrEmailNotVerified   = apperrors.New("EMAIL_NOT_VERIFIED", "Email not verified")
    ErrAccountDisabled   = apperrors.New("ACCOUNT_DISABLED", "Account is disabled")
    ErrTokenExpired      = apperrors.New("TOKEN_EXPIRED", "Token expired")
    ErrTokenRevoked      = apperrors.New("TOKEN_REVOKED", "Token revoked")
    ErrInvalidToken      = apperrors.New("INVALID_TOKEN", "Invalid token")
)

// Register tạo tài khoản mới
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*users.User, error) {
    // Check if email exists
    exists, _ := s.userRepo.ExistsByEmail(ctx, req.Email)
    if exists {
        return nil, users.ErrDuplicateEmail
    }

    // Check if username exists
    exists, _ = s.userRepo.ExistsByUsername(ctx, req.Username)
    if exists {
        return nil, users.ErrDuplicateUsername
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to hash password")
    }

    // Create user (KeyCouple will be generated after email verification)
    user := &users.User{
        ID:            uuid.New(),
        Username:      req.Username,
        Email:         req.Email,
        PasswordHash:  string(hashedPassword),
        Language:      req.Language,
        IsActive:      true,
        EmailVerified: false,
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
    }
    if req.FirstName != nil {
        user.FirstName = req.FirstName
    }
    if req.LastName != nil {
        user.LastName = req.LastName
    }

    // Save user
    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to create user")
    }

    // Gửi mã xác thực email (không block nếu lỗi)
    if err := s.verificationService.CreateVerification(ctx, user.ID, user.Email); err != nil {
        s.log.Warn("Failed to create verification", map[string]interface{}{
            "user_id": user.ID,
            "error":   err.Error(),
        })
    }

    return user, nil
}

// Login authenticates user and returns tokens
func (s *Service) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
    // Get user by email
    user, err := s.userRepo.GetByEmail(ctx, input.Email)
    if err != nil {
        if errors.Is(err, users.ErrNotFound) {
            return nil, ErrInvalidCredentials
        }
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get user")
    }

    // Check password
    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
        return nil, ErrInvalidCredentials
    }

    // Check if user is active
    if !user.IsActive {
        return nil, ErrAccountDisabled
    }

    // Generate access token
    accessToken, err := s.jwtManager.GenerateAccessToken(
        user.ID,
        user.Username,
        user.Email,
        nil, // couple_id will be set separately
    )
    if err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to generate access token")
    }

    // Generate refresh token ID
    refreshTokenID, err := s.jwtManager.GenerateRefreshTokenID()
    if err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to generate refresh token ID")
    }

	// Store refresh token hash
	refreshToken := &domainAuth.RefreshToken{
		ID:          uuid.New(),
		UserID:      user.ID,
		TokenHash:   refreshTokenID, // TODO: hash this before storing
		ExpiresAt:   time.Now().Add(s.refreshTokenTTL),
		DeviceInfo:  &input.DeviceInfo,
		IsRevoked:   false,
		CreatedAt:   time.Now(),
	}

	if err := s.tokenRepo.Create(ctx, refreshToken); err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to store refresh token")
    }

	// Resolve couple status — non-blocking: if this fails we default to false
	couple, _ := s.coupleRepo.GetActiveByUserID(ctx, user.ID)
	hasCouple := couple != nil

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenID,
		ExpiresIn:    int64(s.accessTokenTTL.Seconds()),
		HasCouple:    hasCouple,
		User:         user,
	}, nil
}

// RefreshTokenInput for token refresh
type RefreshTokenInput struct {
    RefreshToken string
    DeviceInfo   string
}

// Refresh exchanges a refresh token for new access token
func (s *Service) Refresh(ctx context.Context, input RefreshTokenInput) (*LoginResult, error) {
    // Get refresh token from DB
    token, err := s.tokenRepo.GetByTokenHash(ctx, input.RefreshToken)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrInvalidToken
        }
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get refresh token")
    }

    // Check validity
    if !token.IsValid() {
        if token.IsExpired() {
            return nil, ErrTokenExpired
        }
        return nil, ErrTokenRevoked
    }

    // Get user
    user, err := s.userRepo.GetByID(ctx, token.UserID)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, apperrors.New("USER_NOT_FOUND", "User not found")
        }
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get user")
    }

    // Generate new access token
    accessToken, err := s.jwtManager.GenerateAccessToken(
        user.ID,
        user.Username,
        user.Email,
        nil,
    )
    if err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to generate access token")
    }

    // Rotate refresh token: revoke old, create new
    if err := s.tokenRepo.Revoke(ctx, token.ID); err != nil {
        // Log but continue
    }

    newRefreshTokenID, err := s.jwtManager.GenerateRefreshTokenID()
    if err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to generate new refresh token ID")
    }

	newRefreshToken := &domainAuth.RefreshToken{
		ID:          uuid.New(),
		UserID:      user.ID,
		TokenHash:   newRefreshTokenID, // TODO: hash
		ExpiresAt:   time.Now().Add(s.refreshTokenTTL),
		DeviceInfo:  &input.DeviceInfo,
		IsRevoked:   false,
		CreatedAt:   time.Now(),
	}

    if err := s.tokenRepo.Create(ctx, newRefreshToken); err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to store new refresh token")
    }

	// Resolve couple status — non-blocking: if this fails we default to false
	couple, _ := s.coupleRepo.GetActiveByUserID(ctx, user.ID)
	hasCouple := couple != nil

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshTokenID,
		ExpiresIn:    int64(s.accessTokenTTL.Seconds()),
		HasCouple:    hasCouple,
		User:         user,
	}, nil
}

// Logout invalidates the refresh token
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
    token, err := s.tokenRepo.GetByTokenHash(ctx, refreshToken)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil // Already logged out
        }
        return apperrors.New("INTERNAL_ERROR", "Failed to get refresh token")
    }

    return s.tokenRepo.Revoke(ctx, token.ID)
}

// LogoutAll invalidates all refresh tokens for user
func (s *Service) LogoutAll(ctx context.Context, userID uuid.UUID) error {
    return s.tokenRepo.RevokeAllForUser(ctx, userID)
}

// GetUserByEmail retrieves a user by email
func (s *Service) GetUserByEmail(ctx context.Context, email string) (*users.User, error) {
    user, err := s.userRepo.GetByEmail(ctx, email)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, users.ErrNotFound
        }
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get user")
    }
    return user, nil
}
