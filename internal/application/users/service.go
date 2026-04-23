package users

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"loviary.app/backend/internal/domain/shared"
	"loviary.app/backend/internal/domain/users"
	"loviary.app/backend/pkg/errors"
)

// Repository defines the interface for user persistence
type Repository interface {
	Create(ctx context.Context, user *users.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*users.User, error)
	GetByEmail(ctx context.Context, email string) (*users.User, error)
	GetByUsername(ctx context.Context, username string) (*users.User, error)
	GetByKeyCouple(ctx context.Context, key string) (*users.User, error)
	Update(ctx context.Context, user *users.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ExistsByKeyCouple(ctx context.Context, key string) (bool, error)
	List(ctx context.Context, limit, offset int) ([]users.User, error)
	Count(ctx context.Context) (int, error)
}

// Service handles user business logic
type Service struct {
	repo Repository
}

// NewService creates a new user service
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreateUserInput represents input for creating a user
type CreateUserInput struct {
	Username    string
	Email       string
	Password    string
	FirstName   *string
	LastName    *string
	DateOfBirth *time.Time
	Gender      *shared.Gender
	Language    string
}

// Create creates a new user
func (s *Service) Create(ctx context.Context, input CreateUserInput) (*users.User, error) {
	// Check if email exists
	exists, err := s.repo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, errors.NewWith("INTERNAL_ERROR", "Failed to check email existence", err)
	}
	if exists {
		return nil, errors.EmailExists
	}

	// Check if username exists
	exists, err = s.repo.ExistsByUsername(ctx, input.Username)
	if err != nil {
		return nil, errors.NewWith("INTERNAL_ERROR", "Failed to check username existence", err)
	}
	if exists {
		return nil, errors.UsernameExists
	}

	hashedPassword, err := hashPassword(input.Password)
	if err != nil {
		return nil, errors.NewWith("INTERNAL_ERROR", "Failed to hash password", err)
	}

	user := &users.User{
		ID:            uuid.New(),
		Username:      input.Username,
		Email:         input.Email,
		PasswordHash:  hashedPassword,
		FirstName:     input.FirstName,
		LastName:      input.LastName,
		DateOfBirth:   input.DateOfBirth,
		Gender:        input.Gender,
		Language:      input.Language,
		IsActive:      true,
		EmailVerified: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, errors.NewWith("INTERNAL_ERROR", "Failed to create user", err)
	}

	return user, nil
}

// UpdateUserInput represents input for updating a user
type UpdateUserInput struct {
	FirstName   *string    `json:"first_name"`
	LastName    *string    `json:"last_name"`
	DateOfBirth *time.Time `json:"date_of_birth"`
	Gender      *shared.Gender `json:"gender"`
	Language    string     `json:"language"`
	AvatarURL   *string    `json:"avatar_url"`
}

// Update updates user information
func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*users.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.UserNotFound
		}
		return nil, errors.NewWith("INTERNAL_ERROR", "Failed to get user", err)
	}

	// Update fields if provided
	if input.FirstName != nil {
		user.FirstName = input.FirstName
	}
	if input.LastName != nil {
		user.LastName = input.LastName
	}
	if input.DateOfBirth != nil {
		user.DateOfBirth = input.DateOfBirth
	}
	if input.Gender != nil {
		user.Gender = input.Gender
	}
	if input.Language != "" {
		user.Language = input.Language
	}
	if input.AvatarURL != nil {
		user.AvatarURL = input.AvatarURL
	}

	user.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, errors.NewWith("INTERNAL_ERROR", "Failed to update user", err)
	}

	return user, nil
}

// UpdateDeviceInput represents input for updating device
type UpdateDeviceInput struct {
	FCMToken   string `json:"fcm_token"`
	Platform   string `json:"platform"`
	DeviceName string `json:"device_name"`
}

// UpdateDevice updates user's FCM device token
func (s *Service) UpdateDevice(ctx context.Context, id uuid.UUID, input UpdateDeviceInput) error {
	// TODO: Implement device update in repository
	// For now, just return nil
	_ = id
	_ = input
	return nil
}

// GetByID retrieves a user by ID
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*users.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.UserNotFound
		}
		return nil, errors.NewWith("INTERNAL_ERROR", "Failed to get user", err)
	}
	return user, nil
}

// GetByKeyCouple retrieves a user by their couple invite key
func (s *Service) GetByKeyCouple(ctx context.Context, key string) (*users.User, error) {
	user, err := s.repo.GetByKeyCouple(ctx, key)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, users.ErrNotFound
		}
		return nil, errors.NewWith("INTERNAL_ERROR", "Failed to get user by key", err)
	}
	return user, nil
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword checks a password against its hash
func VerifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// ToResponse converts a User to a response DTO
func ToResponse(user *users.User) map[string]interface{} {
	return map[string]interface{}{
		"user_id":        user.ID,
		"username":       user.Username,
		"email":          user.Email,
		"first_name":     user.FirstName,
		"last_name":      user.LastName,
		"date_of_birth":  user.DateOfBirth,
		"gender":         user.Gender,
		"language":       user.Language,
		"key_couple":     user.KeyCouple,
		"avatar_url":     user.AvatarURL,
		"is_active":      user.IsActive,
		"email_verified": user.EmailVerified,
		"created_at":     user.CreatedAt,
		"updated_at":     user.UpdatedAt,
	}
}
