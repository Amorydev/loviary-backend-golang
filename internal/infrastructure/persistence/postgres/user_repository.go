package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"loviary.app/backend/internal/domain/users"
	apperrors "loviary.app/backend/pkg/errors"
)

// UserRepository implements users.Repository using PostgreSQL
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user
func (r *UserRepository) Create(ctx context.Context, user *users.User) error {
	const query = `
		INSERT INTO users (
			user_id, username, email, password_hash, first_name, last_name,
			date_of_birth, gender, language, key_couple, avatar_url,
			is_active, email_verified, created_at, updated_at
		) VALUES (
			:user_id, :username, :email, :password_hash, :first_name, :last_name,
			:date_of_birth, :gender, :language, :key_couple, :avatar_url,
			:is_active, :email_verified, :created_at, :updated_at
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" && pqErr.Constraint == "users_email_key" {
				return apperrors.EmailExists
			}
			if pqErr.Code == "23505" && pqErr.Constraint == "users_username_key" {
				return apperrors.UsernameExists
			}
			if pqErr.Code == "23505" && pqErr.Constraint == "users_key_couple_key" {
				return apperrors.New("KEY_COUPLE_EXISTS", "Couple key already exists")
			}
		}
		return apperrors.NewWith("INTERNAL_ERROR", "Failed to create user", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*users.User, error) {
	const query = `
		SELECT user_id, username, email, password_hash, first_name, last_name,
		       date_of_birth, gender, language, key_couple, avatar_url,
		       is_active, email_verified, created_at, updated_at
		FROM users
		WHERE user_id = $1
	`

	var user users.User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*users.User, error) {
	const query = `
		SELECT user_id, username, email, password_hash, first_name, last_name,
		       date_of_birth, gender, language, key_couple, avatar_url,
		       is_active, email_verified, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user users.User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*users.User, error) {
	const query = `
		SELECT user_id, username, email, password_hash, first_name, last_name,
		       date_of_birth, gender, language, key_couple, avatar_url,
		       is_active, email_verified, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	var user users.User
	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByKeyCouple retrieves a user by couple key
func (r *UserRepository) GetByKeyCouple(ctx context.Context, key string) (*users.User, error) {
	const query = `
		SELECT user_id, username, email, password_hash, first_name, last_name,
		       date_of_birth, gender, language, key_couple, avatar_url,
		       is_active, email_verified, created_at, updated_at
		FROM users
		WHERE key_couple = $1
	`

	var user users.User
	err := r.db.GetContext(ctx, &user, query, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.UserNotFound
		}
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to get user by key")
	}
	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *users.User) error {
	const query = `
		UPDATE users
		SET username = :username,
		    email = :email,
		    password_hash = :password_hash,
		    first_name = :first_name,
		    last_name = :last_name,
		    date_of_birth = :date_of_birth,
		    gender = :gender,
		    language = :language,
		    key_couple = :key_couple,
		    avatar_url = :avatar_url,
		    is_active = :is_active,
		    email_verified = :email_verified,
		    updated_at = :updated_at
		WHERE user_id = :user_id
	`

	user.UpdatedAt = time.Now()
	_, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" && pqErr.Constraint == "users_email_key" {
				return apperrors.EmailExists
			}
			if pqErr.Code == "23505" && pqErr.Constraint == "users_username_key" {
				return apperrors.UsernameExists
			}
		}
		return apperrors.NewWith("INTERNAL_ERROR", "Failed to update user", err)
	}
	return nil
}

// Delete removes a user
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM users WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return apperrors.NewWith("INTERNAL_ERROR", "Failed to delete user", err)
	}
	return nil
}

// ExistsByEmail checks if a user with email exists
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	const query = `SELECT 1 FROM users WHERE email = $1 LIMIT 1`
	var exists int
	err := r.db.GetContext(ctx, &exists, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ExistsByUsername checks if a user with username exists
func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	const query = `SELECT 1 FROM users WHERE username = $1 LIMIT 1`
	var exists int
	err := r.db.GetContext(ctx, &exists, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ExistsByKeyCouple checks if a couple key exists
func (r *UserRepository) ExistsByKeyCouple(ctx context.Context, key string) (bool, error) {
	const query = `SELECT 1 FROM users WHERE key_couple = $1 LIMIT 1`
	var exists int
	err := r.db.GetContext(ctx, &exists, query, key)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// List retrieves a paginated list of users
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]users.User, error) {
	const query = `
		SELECT user_id, username, email, password_hash, first_name, last_name,
		       date_of_birth, gender, language, key_couple, avatar_url,
		       is_active, email_verified, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var usersList []users.User
	err := r.db.SelectContext(ctx, &usersList, query, limit, offset)
	if err != nil {
		return nil, err
	}
	return usersList, nil
}

// Count returns the total number of users
func (r *UserRepository) Count(ctx context.Context) (int, error) {
	const query = `SELECT COUNT(*) FROM users`
	var count int
	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, err
	}
	return count, nil
}
