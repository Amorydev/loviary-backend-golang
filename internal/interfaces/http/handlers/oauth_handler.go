package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"loviary.app/backend/internal/application/oauth"
	apperrors "loviary.app/backend/pkg/errors"
)

// OAuthHandler handles OAuth HTTP requests
type OAuthHandler struct {
	oauthService *oauth.Service
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(oauthService *oauth.Service) *OAuthHandler {
	return &OAuthHandler{oauthService: oauthService}
}

// GoogleRedirect redirects user to Google OAuth page
// @Summary Google OAuth redirect
// @Description Redirects to Google OAuth authorization page
// @Tags oauth
// @Produce  json
// @Success  200  {object}  map[string]string "Redirect URL"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /auth/google [get]
func (h *OAuthHandler) GoogleRedirect(c *gin.Context) {
	// Generate a random state parameter for CSRF protection
	// In production, you should store this state in session/Redis and validate it in callback
	state := generateState()

	authURL := h.oauthService.GetAuthURL(state)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"redirect_url": authURL,
		},
	})
}

// GoogleCallback handles the OAuth callback from Google
// @Summary Google OAuth callback
// @Description Handles callback from Google OAuth
// @Tags oauth
// @Accept  json
// @Produce  json
// @Param   code    query    string  true  "Authorization code from Google"
// @Param   state   query    string  false "State parameter for CSRF protection"
// @Success  200  {object}  handlers.LoginResponse "Login successful with tokens"
// @Failure  400  {object}  handlers.ErrorResponse "Missing or invalid code"
// @Failure  401  {object}  handlers.ErrorResponse "OAuth failed"
// @Failure  409  {object}  handlers.ErrorResponse "Email exists with different provider"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /auth/google/callback [get]
func (h *OAuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", "Missing authorization code"))
		return
	}

	// TODO: Validate state parameter for CSRF protection
	_ = state // Use state in production

	// Exchange code for tokens and get user info
	userInfo, err := h.oauthService.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, apperrors.New("OAUTH_FAILED", "OAuth authentication failed"))
		return
	}

	// Login or register user
	result, err := h.oauthService.LoginOrRegister(c.Request.Context(), userInfo)
	if err != nil {
		switch {
		case err.Error() == "EMAIL_EXISTS":
			c.JSON(http.StatusConflict, apperrors.New("EMAIL_EXISTS", "An account with this email already exists"))
		default:
			c.JSON(http.StatusInternalServerError, apperrors.New("INTERNAL_ERROR", "OAuth login failed"))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"access_token":  result.AccessToken,
			"refresh_token": result.RefreshToken,
			"expires_in":    result.ExpiresIn,
			"user": gin.H{
				"user_id":        result.User.ID,
				"username":       result.User.Username,
				"email":          result.User.Email,
				"first_name":     result.User.FirstName,
				"last_name":      result.User.LastName,
				"avatar_url":     result.User.AvatarURL,
				"is_active":      result.User.IsActive,
				"email_verified": result.User.EmailVerified,
			},
		},
	})
}

// GoogleMobile handles OAuth for mobile apps
// Mobile apps typically get the Google ID token directly and send it to backend
// @Summary Google OAuth for mobile
// @Description Exchange Google ID token for app tokens (for mobile apps)
// @Tags oauth
// @Accept  json
// @Produce  json
// @Param   request  body  handlers.GoogleMobileRequest  true  "Google ID token from mobile"
// @Success  200  {object}  handlers.LoginResponse "Login successful with tokens"
// @Failure  400  {object}  handlers.ErrorResponse "Invalid input"
// @Failure  401  {object}  handlers.ErrorResponse "Invalid Google token"
// @Failure  409  {object}  handlers.ErrorResponse "Email exists with different provider"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /auth/google/mobile [post]
func (h *OAuthHandler) GoogleMobile(c *gin.Context) {
	var req GoogleMobileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", err.Error()))
		return
	}

	if req.GoogleToken == "" {
		c.JSON(http.StatusBadRequest, apperrors.New("INVALID_INPUT", "Google token is required"))
		return
	}

	// Verify Google ID token and get user info
	userInfo, err := h.oauthService.VerifyIDToken(c.Request.Context(), req.GoogleToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, apperrors.New("INVALID_GOOGLE_TOKEN", "Invalid Google token"))
		return
	}

	// Login or register user
	result, err := h.oauthService.LoginOrRegister(c.Request.Context(), userInfo)
	if err != nil {
		switch {
		case err.Error() == "EMAIL_EXISTS":
			c.JSON(http.StatusConflict, apperrors.New("EMAIL_EXISTS", "An account with this email already exists"))
		default:
			c.JSON(http.StatusInternalServerError, apperrors.New("INTERNAL_ERROR", "OAuth login failed"))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"access_token":  result.AccessToken,
			"refresh_token": result.RefreshToken,
			"expires_in":    result.ExpiresIn,
			"user": gin.H{
				"user_id":        result.User.ID,
				"username":       result.User.Username,
				"email":          result.User.Email,
				"first_name":     result.User.FirstName,
				"last_name":      result.User.LastName,
				"avatar_url":     result.User.AvatarURL,
				"is_active":      result.User.IsActive,
				"email_verified": result.User.EmailVerified,
			},
		},
	})
}

// GoogleMobileRequest represents mobile OAuth request
type GoogleMobileRequest struct {
	GoogleToken string `json:"google_token" binding:"required"`
}

// generateState generates a random state for CSRF protection
func generateState() string {
	// In production, use crypto/rand to generate a secure random string
	// For now, return a simple string
	return "state_" + http.CanonicalHeaderKey("random")
}
