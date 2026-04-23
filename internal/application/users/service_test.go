package users_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"loviary.app/backend/internal/application/users"
	usersDomain "loviary.app/backend/internal/domain/users"
)

// mockRepository implements users.Repository for testing
type mockRepository struct {
	users map[uuid.UUID]*usersDomain.User
	byEmail map[string]*usersDomain.User
	byUsername map[string]*usersDomain.User
	shouldFail bool
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		users: make(map[uuid.UUID]*usersDomain.User),
		byEmail: make(map[string]*usersDomain.User),
		byUsername: make(map[string]*usersDomain.User),
	}
}

func (m *mockRepository) Create(ctx context.Context, user *usersDomain.User) error {
	if m.shouldFail {
		return errors.New("db error")
	}
	m.users[user.ID] = user
	m.byEmail[user.Email] = user
	m.byUsername[user.Username] = user
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id uuid.UUID) (*usersDomain.User, error) {
	if m.shouldFail {
		return nil, errors.New("db error")
	}
	user, ok := m.users[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return user, nil
}

func (m *mockRepository) GetByEmail(ctx context.Context, email string) (*usersDomain.User, error) {
	if m.shouldFail {
		return nil, errors.New("db error")
	}
	user, ok := m.byEmail[email]
	if !ok {
		return nil, errors.New("not found")
	}
	return user, nil
}

func (m *mockRepository) GetByUsername(ctx context.Context, username string) (*usersDomain.User, error) {
	if m.shouldFail {
		return nil, errors.New("db error")
	}
	user, ok := m.byUsername[username]
	if !ok {
		return nil, errors.New("not found")
	}
	return user, nil
}

func (m *mockRepository) GetByKeyCouple(ctx context.Context, key string) (*usersDomain.User, error) {
	return nil, errors.New("not implemented")
}

func (m *mockRepository) Update(ctx context.Context, user *usersDomain.User) error {
	if m.shouldFail {
		return errors.New("db error")
	}
	m.users[user.ID] = user
	m.byEmail[user.Email] = user
	m.byUsername[user.Username] = user
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.shouldFail {
		return errors.New("db error")
	}
	user, ok := m.users[id]
	if ok {
		delete(m.byEmail, user.Email)
		delete(m.byUsername, user.Username)
		delete(m.users, id)
	}
	return nil
}

func (m *mockRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	if m.shouldFail {
		return false, errors.New("db error")
	}
	_, exists := m.byEmail[email]
	return exists, nil
}

func (m *mockRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	if m.shouldFail {
		return false, errors.New("db error")
	}
	_, exists := m.byUsername[username]
	return exists, nil
}

func (m *mockRepository) ExistsByKeyCouple(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (m *mockRepository) List(ctx context.Context, limit, offset int) ([]usersDomain.User, error) {
	if m.shouldFail {
		return nil, errors.New("db error")
	}
	result := make([]usersDomain.User, 0, len(m.users))
	for _, u := range m.users {
		result = append(result, *u)
	}
	return result, nil
}

func (m *mockRepository) Count(ctx context.Context) (int, error) {
	if m.shouldFail {
		return 0, errors.New("db error")
	}
	return len(m.users), nil
}

func TestUserService_Create(t *testing.T) {
	repo := newMockRepository()
	service := users.NewService(repo)

	tests := []struct {
		name    string
		input   users.CreateUserInput
		wantErr bool
	}{
		{
			name: "successful user creation",
			input: users.CreateUserInput{
				Username: "newuser",
				Email:    "new@example.com",
				Password: "securepassword123",
				FirstName: stringPtr("Test"),
				LastName:  stringPtr("User"),
				Language:  "vi",
			},
			wantErr: false,
		},
		{
			name: "duplicate email",
			input: users.CreateUserInput{
				Username: "testuser2",
				Email:    "test@example.com",
				Password: "securepassword123",
				Language:  "vi",
			},
			wantErr: true,
		},
		{
			name: "duplicate username",
			input: users.CreateUserInput{
				Username: "testuser",
				Email:    "test2@example.com",
				Password: "securepassword123",
				Language:  "vi",
			},
			wantErr: true,
		},
	}

	// Create first user to test duplicates
	_, err := service.Create(context.Background(), users.CreateUserInput{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "securepassword123",
		Language: "vi",
	})
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.Create(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.input.Username, user.Username)
				assert.Equal(t, tt.input.Email, user.Email)
				assert.Equal(t, tt.input.FirstName, user.FirstName)
				assert.Equal(t, tt.input.LastName, user.LastName)
				assert.True(t, user.IsActive)
				assert.False(t, user.EmailVerified)
			}
		})
	}
}

func TestUserService_GetByID(t *testing.T) {
	repo := newMockRepository()
	service := users.NewService(repo)

	// Test user not found
	nonExistentID := uuid.New()
	_, err := service.GetByID(context.Background(), nonExistentID)
	assert.Error(t, err)

	// Create test user
	createdUser, err := service.Create(context.Background(), users.CreateUserInput{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Language: "vi",
	})
	require.NoError(t, err)

	// Test get by ID
	user, err := service.GetByID(context.Background(), createdUser.ID)
	assert.NoError(t, err)
	assert.Equal(t, createdUser.ID, user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)
}

func stringPtr(s string) *string { return &s }
