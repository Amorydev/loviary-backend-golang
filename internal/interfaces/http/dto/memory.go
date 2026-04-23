package dto

import (
    "time"

    "github.com/google/uuid"

    "loviary.app/backend/internal/domain/memories"
)

// MemoryResponse represents memory response DTO
type MemoryResponse struct {
    ID           uuid.UUID             `json:"id"`
    UserID       uuid.UUID             `json:"user_id"`
    CoupleID     *uuid.UUID            `json:"couple_id,omitempty"`
    Title        string                `json:"title"`
    Description  *string               `json:"description,omitempty"`
    MemoryDate   time.Time             `json:"memory_date"`
    MemoryType   memories.MemoryType   `json:"memory_type"`
    MediaURLs    []string              `json:"media_urls,omitempty"`
    Location     *string               `json:"location,omitempty"`
    IsPrivate    bool                  `json:"is_private"`
    IsShared     bool                  `json:"is_shared"`
    CreatedAt    time.Time             `json:"created_at"`
    UpdatedAt    time.Time             `json:"updated_at"`
}

// MemoryToResponse converts a Memory to response DTO
func MemoryToResponse(memory *memories.Memory) MemoryResponse {
    return MemoryResponse{
        ID:          memory.ID,
        UserID:      memory.UserID,
        CoupleID:    memory.CoupleID,
        Title:       memory.Title,
        Description: memory.Description,
        MemoryDate:  memory.MemoryDate,
        MemoryType:  memory.MemoryType,
        MediaURLs:   memory.MediaURLs,
        Location:    memory.Location,
        IsPrivate:   memory.IsPrivate,
        IsShared:    memory.IsShared,
        CreatedAt:   memory.CreatedAt,
        UpdatedAt:   memory.UpdatedAt,
    }
}
