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
	InviteKey        string                  `json:"invite_key" binding:"required"`
	CoupleName       *string                 `json:"couple_name"`
	RelationshipType shared.RelationshipType `json:"relationship_type" binding:"omitempty,oneof=dating engaged married"`
}

// InviteResponse represents invite response
type InviteResponse struct {
	CoupleID            uuid.UUID           `json:"couple_id"`
	Status              shared.CoupleStatus `json:"status"`
	InvitationExpiresAt *time.Time          `json:"invitation_expires_at,omitempty"`
	Message             string              `json:"message"`
}

// Invite creates a couple invitation using partner's invite key
// @Summary Create couple invitation
// @Description Send a couple invitation to another user using their invite key
// @Tags couples
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param   request  body  handlers.InviteRequest  true  "Invite request with partner's invite key"
// @Success  201  {object}  handlers.InviteResponse "Invitation created"
// @Failure  400  {object}  handlers.ErrorResponse "Invalid input"
// @Failure  404  {object}  handlers.ErrorResponse "User with invite key not found"
// @Failure  409  {object}  handlers.ErrorResponse "Invitation conflict"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /couples/invite [post]
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
		User1ID:          userID,
		User2ID:          partnerUser.ID,
		CoupleName:       req.CoupleName,
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
		CoupleID:            couple.CoupleID,
		Status:              couple.Status,
		InvitationExpiresAt: couple.InvitationExpiresAt,
		Message:             message,
	})
}

// GetMyCouple retrieves the authenticated user's couple
// @Summary Get my couple
// @Description Get current couple relationship information
// @Tags couples
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success  200  {object}  handlers.CoupleResponse "Couple data"
// @Failure  404  {object}  handlers.ErrorResponse "No active couple found"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /couples/me [get]
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
// @Summary Confirm couple invitation
// @Description Accept a couple invitation sent to the user
// @Tags couples
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Param   request  body  handlers.ConfirmInvitationRequest  true  "Confirmation request with couple ID"
// @Success  200  {object}  handlers.MessageResponse "Invitation accepted"
// @Failure  400  {object}  handlers.ErrorResponse "Invalid input"
// @Failure  401  {object}  handlers.ErrorResponse "Not authenticated"
// @Failure  404  {object}  handlers.ErrorResponse "Couple not found"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /couples/confirm [post]
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
// @Summary Delete my couple
// @Description Initiate breakup of current couple relationship (grace period applies)
// @Tags couples
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success  200  {object}  handlers.MessageResponse "Breakup initiated"
// @Failure  401  {object}  handlers.ErrorResponse "Not authenticated"
// @Failure  404  {object}  handlers.ErrorResponse "No active couple found"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /couples/me [delete]
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
