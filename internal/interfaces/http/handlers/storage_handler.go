package handlers

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"

    "loviary.app/backend/internal/interfaces/http/middleware"
    apperrors "loviary.app/backend/pkg/errors"
)

// StorageHandler handles storage-related HTTP requests
type StorageHandler struct{}

// NewStorageHandler creates a new storage handler
func NewStorageHandler() *StorageHandler {
    return &StorageHandler{}
}

// PresignRequest represents presign URL request
type PresignRequest struct {
    FileName    string `json:"file_name" binding:"required"`
    ContentType string `json:"content_type" binding:"required"`
    FileSize    int64  `json:"file_size" binding:"required,min=1"`
}

// PresignResponse represents presign URL response
type PresignResponse struct {
    UploadURL  string    `json:"upload_url"`
    FileURL    string    `json:"file_url"`
    ExpiresAt  time.Time `json:"expires_at"`
    SignedID   string    `json:"signed_id"`
}

// Presign generates a presigned URL for file upload
// @Summary Generate presigned upload URL
// @Description Get a presigned URL for uploading files to storage
// @Tags storage
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param   request  body  handlers.PresignRequest  true  "Presign request"
// @Success  200  {object}  handlers.PresignResponse "Presigned URL and file info"
// @Failure  400  {object}  handlers.ErrorResponse "Invalid input or file too large"
// @Failure  401  {object}  handlers.ErrorResponse "Not authenticated"
// @Router   /storage/presign [post]
func (h *StorageHandler) Presign(c *gin.Context) {
    userID, exists := middleware.GetUserID(c)
    if !exists {
        c.JSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
        return
    }

    var req PresignRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", err.Error()))
        return
    }

    // Validate file size (max 10MB)
    if req.FileSize > 10*1024*1024 {
        c.JSON(http.StatusBadRequest, apperrors.New("INVALID_FILE_SIZE", "File size exceeds 10MB limit"))
        return
    }

    // Generate a unique file key
    fileKey := "uploads/" + userID.String() + "/" + uuid.New().String() + "_" + req.FileName

    // In a real implementation, this would generate a presigned URL from S3 or similar
    // For now, return a placeholder response
    expiresAt := time.Now().Add(1 * time.Hour)

    c.JSON(http.StatusOK, PresignResponse{
        UploadURL: "https://storage.example.com/upload?key=" + fileKey,
        FileURL:   "https://storage.example.com/" + fileKey,
        ExpiresAt: expiresAt,
        SignedID:  fileKey,
    })
}
