package memories

import (
    "context"
    "errors"
    "time"

    "github.com/google/uuid"

    "loviary.app/backend/internal/domain/couples"
    "loviary.app/backend/internal/domain/memories"
    "loviary.app/backend/internal/domain/users"
    apperrors "loviary.app/backend/pkg/errors"
)

// Repository defines the interface for memory persistence
type Repository interface {
    Create(ctx context.Context, memory *memories.Memory) error
    GetByID(ctx context.Context, id uuid.UUID) (*memories.Memory, error)
    GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]memories.Memory, error)
    GetByCoupleID(ctx context.Context, coupleID uuid.UUID, limit, offset int) ([]memories.Memory, error)
    Update(ctx context.Context, memory *memories.Memory) error
    Delete(ctx context.Context, id uuid.UUID) error
    CountByUserID(ctx context.Context, userID uuid.UUID) (int, error)
}

// Service handles memory business logic
type Service struct {
    repo      Repository
    userRepo  interface{ GetByID(ctx context.Context, id uuid.UUID) (*users.User, error) }
    coupleRepo interface{ GetByID(ctx context.Context, id uuid.UUID) (*couples.Couple, error) }
}

// NewService creates a new memory service
func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}

// CreateMemoryInput represents input for creating a memory
type CreateMemoryInput struct {
    UserID      uuid.UUID
    CoupleID    *uuid.UUID
    Title       string
    Description *string
    MemoryDate  time.Time
    MemoryType  memories.MemoryType
    MediaURLs   []string
    Location    *string
    IsPrivate   bool
    IsShared    bool
}

// Create creates a new memory
func (s *Service) Create(ctx context.Context, input CreateMemoryInput) (*memories.Memory, error) {
    // Validate memory type
    if !input.MemoryType.IsValid() {
        return nil, apperrors.New("INVALID_MEMORY_TYPE", "Invalid memory type")
    }

    // If couple_id is provided, verify the user is part of that couple
    if input.CoupleID != nil {
        // TODO: Verify couple membership when couple repo is available
    }

    now := time.Now()
    memory := &memories.Memory{
        ID:           uuid.New(),
        UserID:       input.UserID,
        CoupleID:     input.CoupleID,
        Title:        input.Title,
        Description:  input.Description,
        MemoryDate:   input.MemoryDate,
        MemoryType:   input.MemoryType,
        MediaURLs:    input.MediaURLs,
        Location:     input.Location,
        IsPrivate:    input.IsPrivate,
        IsShared:     input.IsShared,
        CreatedAt:    now,
        UpdatedAt:    now,
    }

    if err := s.repo.Create(ctx, memory); err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to create memory")
    }

    return memory, nil
}

// UpdateMemoryInput represents input for updating a memory
type UpdateMemoryInput struct {
    Title        *string
    Description  *string
    MemoryDate   *time.Time
    MemoryType   *memories.MemoryType
    MediaURLs    *[]string
    Location     *string
    IsPrivate    *bool
    IsShared     *bool
}

// Update updates an existing memory
func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateMemoryInput) (*memories.Memory, error) {
    memory, err := s.repo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, apperrors.New("NOT_FOUND", "")) {
            return nil, apperrors.New("MEMORY_NOT_FOUND", "Memory not found")
        }
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get memory")
    }

    // Update fields if provided
    if input.Title != nil {
        memory.Title = *input.Title
    }
    if input.Description != nil {
        memory.Description = input.Description
    }
    if input.MemoryDate != nil {
        memory.MemoryDate = *input.MemoryDate
    }
    if input.MemoryType != nil {
        if !(*input.MemoryType).IsValid() {
            return nil, apperrors.New("INVALID_MEMORY_TYPE", "Invalid memory type")
        }
        memory.MemoryType = *input.MemoryType
    }
    if input.MediaURLs != nil {
        memory.MediaURLs = *input.MediaURLs
    }
    if input.Location != nil {
        memory.Location = input.Location
    }
    if input.IsPrivate != nil {
        memory.IsPrivate = *input.IsPrivate
    }
    if input.IsShared != nil {
        memory.IsShared = *input.IsShared
    }

    memory.UpdatedAt = time.Now()

    if err := s.repo.Update(ctx, memory); err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to update memory")
    }

    return memory, nil
}

// GetMemory retrieves a memory by ID
func (s *Service) GetMemory(ctx context.Context, id uuid.UUID) (*memories.Memory, error) {
    memory, err := s.repo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, apperrors.New("NOT_FOUND", "")) {
            return nil, apperrors.New("MEMORY_NOT_FOUND", "Memory not found")
        }
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get memory")
    }
    return memory, nil
}

// GetMemoriesByUser retrieves memories for a user
func (s *Service) GetMemoriesByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]memories.Memory, error) {
    if limit <= 0 {
        limit = 20 // default limit
    }
    memories, err := s.repo.GetByUserID(ctx, userID, limit, offset)
    if err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get memories")
    }
    return memories, nil
}

// GetMemoriesByCouple retrieves shared memories for a couple
func (s *Service) GetMemoriesByCouple(ctx context.Context, coupleID uuid.UUID, limit, offset int) ([]memories.Memory, error) {
    if limit <= 0 {
        limit = 20 // default limit
    }
    memories, err := s.repo.GetByCoupleID(ctx, coupleID, limit, offset)
    if err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get couple memories")
    }
    return memories, nil
}

// DeleteMemory removes a memory
func (s *Service) DeleteMemory(ctx context.Context, id uuid.UUID) error {
    // First verify the memory exists
    _, err := s.repo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, apperrors.MemoryNotFound) {
            return apperrors.New("MEMORY_NOT_FOUND", "Memory not found")
        }
        return apperrors.New("INTERNAL_ERROR", "Failed to get memory")
    }

    if err := s.repo.Delete(ctx, id); err != nil {
        return apperrors.New("INTERNAL_ERROR", "Failed to delete memory")
    }
    return nil
}

// CountMemories returns the count of memories for a user
func (s *Service) CountMemories(ctx context.Context, userID uuid.UUID) (int, error) {
    count, err := s.repo.CountByUserID(ctx, userID)
    if err != nil {
        return 0, apperrors.New("INTERNAL_ERROR", "Failed to count memories")
    }
    return count, nil
}
