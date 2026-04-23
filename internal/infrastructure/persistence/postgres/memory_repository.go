package persistence

import (
    "context"
    "database/sql"
    "errors"

    "github.com/google/uuid"

    "loviary.app/backend/internal/domain/memories"
    apperrors "loviary.app/backend/pkg/errors"
)

// MemoryRepository handles database operations for memories
type MemoryRepository struct {
    db *sql.DB
}

// NewMemoryRepository creates a new memory repository
func NewMemoryRepository(db *sql.DB) *MemoryRepository {
    return &MemoryRepository{db: db}
}

// Create inserts a new memory
func (r *MemoryRepository) Create(ctx context.Context, memory *memories.Memory) error {
    query := `
        INSERT INTO memories (id, user_id, couple_id, title, description, memory_date, memory_type, media_urls, location, is_private, is_shared, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
    `
    _, err := r.db.ExecContext(ctx, query,
        memory.ID,
        memory.UserID,
        memory.CoupleID,
        memory.Title,
        memory.Description,
        memory.MemoryDate,
        memory.MemoryType,
        memory.MediaURLs,
        memory.Location,
        memory.IsPrivate,
        memory.IsShared,
        memory.CreatedAt,
        memory.UpdatedAt,
    )
    if err != nil {
        return apperrors.New("INTERNAL_ERROR", "Failed to create memory")
    }
    return nil
}

// GetByID retrieves a memory by ID
func (r *MemoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*memories.Memory, error) {
    query := `
        SELECT id, user_id, couple_id, title, description, memory_date, memory_type, media_urls, location, is_private, is_shared, created_at, updated_at
        FROM memories
        WHERE id = $1
    `
    var memory memories.Memory
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &memory.ID,
        &memory.UserID,
        &memory.CoupleID,
        &memory.Title,
        &memory.Description,
        &memory.MemoryDate,
        &memory.MemoryType,
        &memory.MediaURLs,
        &memory.Location,
        &memory.IsPrivate,
        &memory.IsShared,
        &memory.CreatedAt,
        &memory.UpdatedAt,
    )
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, apperrors.MemoryNotFound
        }
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to get memory")
    }
    return &memory, nil
}

// GetByUserID retrieves all memories for a user
func (r *MemoryRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]memories.Memory, error) {
    query := `
        SELECT id, user_id, couple_id, title, description, memory_date, memory_type, media_urls, location, is_private, is_shared, created_at, updated_at
        FROM memories
        WHERE user_id = $1
        ORDER BY memory_date DESC
        LIMIT $2 OFFSET $3
    `
    rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
    if err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to query memories")
    }
    defer rows.Close()

    var memoryList []memories.Memory
    for rows.Next() {
        var memory memories.Memory
        if err := rows.Scan(
            &memory.ID,
            &memory.UserID,
            &memory.CoupleID,
            &memory.Title,
            &memory.Description,
            &memory.MemoryDate,
            &memory.MemoryType,
            &memory.MediaURLs,
            &memory.Location,
            &memory.IsPrivate,
            &memory.IsShared,
            &memory.CreatedAt,
            &memory.UpdatedAt,
        ); err != nil {
            return nil, apperrors.New("INTERNAL_ERROR", "Failed to scan memory row")
        }
        memoryList = append(memoryList, memory)
    }
    if err := rows.Err(); err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to iterate memory rows")
    }
    return memoryList, nil
}

// GetByCoupleID retrieves shared memories for a couple
func (r *MemoryRepository) GetByCoupleID(ctx context.Context, coupleID uuid.UUID, limit, offset int) ([]memories.Memory, error) {
    query := `
        SELECT id, user_id, couple_id, title, description, memory_date, memory_type, media_urls, location, is_private, is_shared, created_at, updated_at
        FROM memories
        WHERE couple_id = $1 AND is_shared = true
        ORDER BY memory_date DESC
        LIMIT $2 OFFSET $3
    `
    rows, err := r.db.QueryContext(ctx, query, coupleID, limit, offset)
    if err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to query couple memories")
    }
    defer rows.Close()

    var memoryList []memories.Memory
    for rows.Next() {
        var memory memories.Memory
        if err := rows.Scan(
            &memory.ID,
            &memory.UserID,
            &memory.CoupleID,
            &memory.Title,
            &memory.Description,
            &memory.MemoryDate,
            &memory.MemoryType,
            &memory.MediaURLs,
            &memory.Location,
            &memory.IsPrivate,
            &memory.IsShared,
            &memory.CreatedAt,
            &memory.UpdatedAt,
        ); err != nil {
            return nil, apperrors.New("INTERNAL_ERROR", "Failed to scan memory row")
        }
        memoryList = append(memoryList, memory)
    }
    if err := rows.Err(); err != nil {
        return nil, apperrors.New("INTERNAL_ERROR", "Failed to iterate memory rows")
    }
    return memoryList, nil
}

// Update updates an existing memory
func (r *MemoryRepository) Update(ctx context.Context, memory *memories.Memory) error {
    query := `
        UPDATE memories
        SET title = $2, description = $3, memory_date = $4, memory_type = $5,
            media_urls = $6, location = $7, is_private = $8, is_shared = $9, updated_at = $10
        WHERE id = $1
    `
    _, err := r.db.ExecContext(ctx, query,
        memory.ID,
        memory.Title,
        memory.Description,
        memory.MemoryDate,
        memory.MemoryType,
        memory.MediaURLs,
        memory.Location,
        memory.IsPrivate,
        memory.IsShared,
        memory.UpdatedAt,
    )
    if err != nil {
        return apperrors.New("INTERNAL_ERROR", "Failed to update memory")
    }
    return nil
}

// Delete removes a memory
func (r *MemoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
    query := `DELETE FROM memories WHERE id = $1`
    result, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        return apperrors.New("INTERNAL_ERROR", "Failed to delete memory")
    }
    if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
        return apperrors.MemoryNotFound
    }
    return nil
}

// CountByUserID returns the count of memories for a user
func (r *MemoryRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
    query := `SELECT COUNT(*) FROM memories WHERE user_id = $1`
    var count int
    err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
    if err != nil {
        return 0, apperrors.New("INTERNAL_ERROR", "Failed to count memories")
    }
    return count, nil
}
