package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	appMemories "loviary.app/backend/internal/application/memories"
	"loviary.app/backend/internal/domain/memories"
	"loviary.app/backend/internal/interfaces/http/dto"
	"loviary.app/backend/internal/interfaces/http/middleware"
	apperrors "loviary.app/backend/pkg/errors"
)

// MemoryHandler handles memory-related HTTP requests
type MemoryHandler struct {
	memoryService *appMemories.Service
}

// NewMemoryHandler creates a new memory handler
func NewMemoryHandler(memoryService *appMemories.Service) *MemoryHandler {
	return &MemoryHandler{memoryService: memoryService}
}

// CreateMemoryRequest represents create memory request
type CreateMemoryRequest struct {
	CoupleID    *uuid.UUID          `json:"couple_id"`
	Title       string              `json:"title" binding:"required"`
	Description *string             `json:"description"`
	MemoryDate  time.Time           `json:"memory_date" binding:"required"`
	MemoryType  memories.MemoryType `json:"memory_type" binding:"required,oneof=date_night trip milestone everyday celebration achievement"`
	MediaURLs   []string            `json:"media_urls"`
	Location    *string             `json:"location"`
	IsPrivate   bool                `json:"is_private"`
	IsShared    bool                `json:"is_shared"`
}

// UpdateMemoryRequest represents update memory request
type UpdateMemoryRequest struct {
	Title       *string              `json:"title"`
	Description *string              `json:"description"`
	MemoryDate  *time.Time           `json:"memory_date"`
	MemoryType  *memories.MemoryType `json:"memory_type" binding:"omitempty,oneof=date_night trip milestone everyday celebration achievement"`
	MediaURLs   *[]string            `json:"media_urls"`
	Location    *string              `json:"location"`
	IsPrivate   *bool                `json:"is_private"`
	IsShared    *bool                `json:"is_shared"`
}

// CreateMemory creates a new memory
// @Summary Create memory
// @Description Create a new memory (special moment) entry
// @Tags memories
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param   request  body  handlers.CreateMemoryRequest  true  "Memory data"
// @Success  201  {object}  dto.MemoryResponse "Memory created"
// @Failure  400  {object}  handlers.ErrorResponse "Invalid input"
// @Failure  401  {object}  handlers.ErrorResponse "Not authenticated"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /memories [post]
func (h *MemoryHandler) CreateMemory(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
		return
	}

	var req CreateMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", err.Error()))
		return
	}

	memory, err := h.memoryService.Create(c.Request.Context(), appMemories.CreateMemoryInput{
		UserID:      userID,
		CoupleID:    req.CoupleID,
		Title:       req.Title,
		Description: req.Description,
		MemoryDate:  req.MemoryDate,
		MemoryType:  req.MemoryType,
		MediaURLs:   req.MediaURLs,
		Location:    req.Location,
		IsPrivate:   req.IsPrivate,
		IsShared:    req.IsShared,
	})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, dto.MemoryToResponse(memory))
}

// GetMemory retrieves a memory by ID
// @Summary Get memory
// @Description Get memory by ID
// @Tags memories
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param   id  path  string  true  "Memory ID (UUID)"
// @Success  200  {object}  dto.MemoryResponse "Memory data"
// @Failure  400  {object}  handlers.ErrorResponse "Invalid memory ID"
// @Failure  401  {object}  handlers.ErrorResponse "Not authenticated"
// @Failure  404  {object}  handlers.ErrorResponse "Memory not found"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /memories/{id} [get]
func (h *MemoryHandler) GetMemory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", "Invalid memory ID"))
		return
	}

	memory, err := h.memoryService.GetMemory(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.MemoryToResponse(memory))
}

// GetMemories retrieves memories for the authenticated user
// @Summary Get memories
// @Description Get memories for the current user with pagination
// @Tags memories
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param   limit  query  int  false  "Number of items per page (max 100, default 20)"
// @Param   offset  query  int  false  "Number of items to skip (default 0)"
// @Success  200  {object}  handlers.MemoryListResponse "List of memories with pagination info"
// @Failure  400  {object}  handlers.ErrorResponse "Invalid pagination parameters"
// @Failure  401  {object}  handlers.ErrorResponse "Not authenticated"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /memories [get]
func (h *MemoryHandler) GetMemories(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
		return
	}

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit := 20
	if _, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || limit <= 0 || limit > 100 {
		c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", "Invalid limit"))
		return
	}
	offset := 0
	if _, err := fmt.Sscanf(offsetStr, "%d", &offset); err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", "Invalid offset"))
		return
	}

	memories, err := h.memoryService.GetMemoriesByUser(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.Error(err)
		return
	}

	response := make([]dto.MemoryResponse, len(memories))
	for i, memory := range memories {
		response[i] = dto.MemoryToResponse(&memory)
	}

	c.JSON(http.StatusOK, gin.H{
		"memories": response,
		"count":    len(response),
		"limit":    limit,
		"offset":   offset,
	})
}

// UpdateMemory updates a memory
// @Summary Update memory
// @Description Update an existing memory entry
// @Tags memories
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param   id  path  string  true  "Memory ID (UUID)"
// @Param   request  body  handlers.UpdateMemoryRequest  true  "Updated memory data"
// @Success  200  {object}  dto.MemoryResponse "Updated memory"
// @Failure  400  {object}  handlers.ErrorResponse "Invalid input"
// @Failure  401  {object}  handlers.ErrorResponse "Not authenticated"
// @Failure  404  {object}  handlers.ErrorResponse "Memory not found"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /memories/{id} [patch]
func (h *MemoryHandler) UpdateMemory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", "Invalid memory ID"))
		return
	}

	var req UpdateMemoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", err.Error()))
		return
	}

	memory, err := h.memoryService.Update(c.Request.Context(), id, appMemories.UpdateMemoryInput{
		Title:       req.Title,
		Description: req.Description,
		MemoryDate:  req.MemoryDate,
		MemoryType:  req.MemoryType,
		MediaURLs:   req.MediaURLs,
		Location:    req.Location,
		IsPrivate:   req.IsPrivate,
		IsShared:    req.IsShared,
	})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, dto.MemoryToResponse(memory))
}

// DeleteMemory removes a memory
// @Summary Delete memory
// @Description Delete a memory by ID
// @Tags memories
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param   id  path  string  true  "Memory ID (UUID)"
// @Success  200  {object}  handlers.SuccessResponse "Memory deleted"
// @Failure  400  {object}  handlers.ErrorResponse "Invalid memory ID"
// @Failure  401  {object}  handlers.ErrorResponse "Not authenticated"
// @Failure  404  {object}  handlers.ErrorResponse "Memory not found"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /memories/{id} [delete]
func (h *MemoryHandler) DeleteMemory(c *gin.Context) {
	if _, exists := middleware.GetUserID(c); !exists {
		c.JSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", "Invalid memory ID"))
		return
	}

	if err := h.memoryService.DeleteMemory(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
