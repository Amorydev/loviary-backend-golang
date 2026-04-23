package handlers

import (
    "errors"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"

    "loviary.app/backend/internal/application/couples"
    appUsers "loviary.app/backend/internal/application/users"
    "loviary.app/backend/internal/domain/shared"
    users "loviary.app/backend/internal/domain/users"
    "loviary.app/backend/internal/interfaces/http/middleware"
    apperrors "loviary.app/backend/pkg/errors"
)
// CoupleHandler handles couple-related HTTP requests
type CoupleHandler struct {
    coupleService *couples.Service
    userService   *appUsers.Service
}

// NewCoupleHandler creates a new couple handler
func NewCoupleHandler(coupleService *couples.Service, userService *appUsers.Service) *CoupleHandler {
    return &CoupleHandler{
        coupleService: coupleService,
        userService:   userService,
    }
}

// InviteRequest represents invite couple request
type InviteRequest struct {
	InviteKey       string                 `json:"invite_key" binding:"required"`
	CoupleName      *string                `json:"couple_name"`
	RelationshipType shared.RelationshipType `json:"relationship_type" binding:"omitempty,oneof=dating engaged married"`
}

// InviteResponse represents invite response
type InviteResponse struct {
	CoupleID        uuid.UUID              `json:"couple_id"`
	Status          shared.CoupleStatus    `json:"status"`
	InvitationExpiresAt *time.Time          `json:"invitation_expires_at,omitempty"`
	Message         string                 `json:"message"`
}

// Invite creates a couple invitation using partner's invite key
// POST /api/v1/couples/invite
func (h *CoupleHandler) Invite(c *gin.Context) {
	// Get authenticated user ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
		return
	}

	var req InviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", err.Error()))
		return
	}

	// Look up partner user by invite key
	partnerUser, err := h.userService.GetByKeyCouple(c.Request.Context(), req.InviteKey)
	if err != nil {
		if err == users.ErrNotFound {
			c.JSON(http.StatusNotFound, apperrors.New("USER_NOT_FOUND", "User with this invite key not found"))
			return
		}
		c.Error(err)
		return
	}

	couple, err := h.coupleService.Create(c.Request.Context(), couples.CreateCoupleInput{
		User1ID:    userID,
		User2ID:    partnerUser.ID,
		CoupleName: req.CoupleName,
		RelationshipType: req.RelationshipType,
	})
	if err != nil {
		c.Error(err)
		return
	}

	message := "Invitation sent successfully"
	if couple.Status == shared.CoupleStatusActive {
		message = "Couple created successfully"
	}

	c.JSON(http.StatusCreated, InviteResponse{
		CoupleID: couple.CoupleID,
		Status: couple.Status,
		InvitationExpiresAt: couple.InvitationExpiresAt,
		Message: message,
	})
}

// GetMyCouple retrieves the authenticated user's couple
// GET /api/v1/couples/me
func (h *CoupleHandler) GetMyCouple(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
		return
	}

	couple, err := h.coupleService.GetActiveByUserID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, apperrors.NoActiveCouple) {
			c.JSON(http.StatusNotFound, apperrors.NoActiveCouple)
			return
		}
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, couple)
}

// ConfirmInvitationRequest represents confirm invitation request
type ConfirmInvitationRequest struct {
	CoupleID uuid.UUID `json:"couple_id" binding:"required"`
}

// ConfirmInvitation accepts a couple invitation
// POST /api/v1/couples/confirm
func (h *CoupleHandler) ConfirmInvitation(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
		return
	}

	var req ConfirmInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", err.Error()))
		return
	}

	if err := h.coupleService.AcceptInvitation(c.Request.Context(), req.CoupleID, userID); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invitation accepted successfully"})
}

// DeleteMyCouple deletes the authenticated user's couple
// DELETE /api/v1/couples/me
func (h *CoupleHandler) DeleteMyCouple(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
		return
	}

	// Get user's active couple
	couple, err := h.coupleService.GetActiveByUserID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, apperrors.NoActiveCouple) {
			c.JSON(http.StatusNotFound, apperrors.NoActiveCouple)
			return
		}
		c.Error(err)
		return
	}

	// Initiate breakup (will be confirmed after grace period)
	if err := h.coupleService.InitiateBreakup(c.Request.Context(), couple.CoupleID, userID); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Breakup initiated. The couple will be permanently deleted after the grace period."})
}
