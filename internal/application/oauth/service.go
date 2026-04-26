package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	appAuth "loviary.app/backend/internal/application/auth"
	domainAuth "loviary.app/backend/internal/domain/auth"
	"loviary.app/backend/internal/domain/users"
	postgres "loviary.app/backend/internal/infrastructure/persistence/postgres"
	"loviary.app/backend/internal/infrastructure/jwt"
	apperrors "loviary.app/backend/pkg/errors"
)

// Service handles Google OAuth authentication
type Service struct {
	userRepo    *postgres.UserRepository
	tokenRepo   *postgres.RefreshTokenRepository
	jwtManager  *jwt.Manager
	oauthConfig *oauth2.Config
	accessTTL   time.Duration
	refreshTTL  time.Duration
}

// GoogleUserInfo represents the user info from Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	FirstName     string `json:"given_name"`
	LastName      string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// NewService creates a new OAuth service
func NewService(
	userRepo *postgres.UserRepository,
	tokenRepo *postgres.RefreshTokenRepository,
	jwtManager *jwt.Manager,
	googleClientID string,
	googleClientSecret string,
	redirectURI string,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *Service {
	oauthConfig := &oauth2.Config{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		RedirectURL:  redirectURI,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &Service{
		userRepo:    userRepo,
		tokenRepo:   tokenRepo,
		jwtManager:  jwtManager,
		oauthConfig: oauthConfig,
		accessTTL:   accessTTL,
		refreshTTL:  refreshTTL,
	}
}

// GetAuthURL returns the Google OAuth URL to redirect the user to
func (s *Service) GetAuthURL(state string) string {
	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges authorization code for tokens and user info
func (s *Service) ExchangeCode(ctx context.Context, code string) (*GoogleUserInfo, error) {
	// Exchange code for token
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, apperrors.New("OAUTH_EXCHANGE_FAILED", "Failed to exchange code for token")
	}

	// Get user info from Google
	client := s.oauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, apperrors.New("OAUTH_USERINFO_FAILED", "Failed to get user info from Google")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apperrors.New("OAUTH_USERINFO_ERROR", "Google returned error status")
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, apperrors.New("OAUTH_PARSE_FAILED", "Failed to parse user info")
	}

	return &userInfo, nil
}

// VerifyIDToken verifies a Google ID token from mobile and returns user info
func (s *Service) VerifyIDToken(ctx context.Context, idToken string) (*GoogleUserInfo, error) {
	// Verify the ID token using Google's token info endpoint
	verificationURL := "https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(idToken)

	resp, err := http.Get(verificationURL)
	if err != nil {
		return nil, apperrors.New("TOKEN_VERIFICATION_FAILED", "Failed to verify token")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apperrors.New("INVALID_TOKEN", "Invalid Google token")
	}

	var claims map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return nil, apperrors.New("TOKEN_PARSE_FAILED", "Failed to parse token claims")
	}

	// Verify the audience (client ID)
	if aud, ok := claims["aud"].(string); ok {
		if aud != s.oauthConfig.ClientID {
			return nil, apperrors.New("TOKEN_INVALID_CLIENT", "Token was not issued for this client")
		}
	} else {
		// Multiple audiences possible, check array
		if audiences, ok := claims["aud"].([]interface{}); ok {
			found := false
			for _, aud := range audiences {
				if audStr, ok := aud.(string); ok && audStr == s.oauthConfig.ClientID {
					found = true
					break
				}
			}
			if !found {
				return nil, apperrors.New("TOKEN_INVALID_CLIENT", "Token was not issued for this client")
			}
		}
	}

	// Extract user info from claims
	userInfo := &GoogleUserInfo{
		ID:            extractString(claims, "sub"),
		Email:         extractString(claims, "email"),
		VerifiedEmail: extractBool(claims, "email_verified"),
		FirstName:     extractString(claims, "given_name"),
		LastName:      extractString(claims, "family_name"),
		Picture:       extractString(claims, "picture"),
		Locale:        extractString(claims, "locale"),
	}

	if userInfo.ID == "" || userInfo.Email == "" {
		return nil, apperrors.New("INVALID_TOKEN", "Token missing required claims")
	}

	return userInfo, nil
}

// Helper functions for extracting claims
func extractString(claims map[string]interface{}, key string) string {
	if val, ok := claims[key].(string); ok {
		return val
	}
	return ""
}

func extractBool(claims map[string]interface{}, key string) bool {
	if val, ok := claims[key].(bool); ok {
		return val
	}
	if val, ok := claims[key].(string); ok {
		return val == "true" || val == "1"
	}
	return false
}

// LoginOrRegister handles OAuth callback - either finds existing user or creates new one
func (s *Service) LoginOrRegister(ctx context.Context, userInfo *GoogleUserInfo) (*appAuth.LoginResult, error) {
	// Check if user exists with this Google ID
	existingUser, err := s.userRepo.GetByProvider(ctx, "google", userInfo.ID)
	if err == nil {
		// Existing OAuth user - update their info and login
		existingUser.FirstName = &userInfo.FirstName
		existingUser.LastName = &userInfo.LastName
		existingUser.AvatarURL = &userInfo.Picture
		existingUser.EmailVerified = userInfo.VerifiedEmail

		// Update user
		if updateErr := s.userRepo.Update(ctx, existingUser); updateErr != nil {
			return nil, apperrors.New("INTERNAL_ERROR", "Failed to update user")
		}

		// Generate tokens
		return s.generateTokens(ctx, existingUser)
	}

	// Check if user exists with this email (regular account)
	_, emailErr := s.userRepo.GetByEmail(ctx, userInfo.Email)
	if emailErr == nil {
		// User exists with email but not linked to Google
		return nil, apperrors.New("EMAIL_EXISTS", "An account with this email already exists. Please login with password and link Google in settings.")
	}

	// Create new user from Google info
	newUser := &users.User{
		ID:              uuid.New(),
		Username:        generateUsername(userInfo),
		Email:           userInfo.Email,
		FirstName:       &userInfo.FirstName,
		LastName:       &userInfo.LastName,
		AvatarURL:       &userInfo.Picture,
		IsActive:        true,
		EmailVerified:   userInfo.VerifiedEmail,
		AuthProvider:    stringPtr("google"),
		AuthProviderID:  &userInfo.ID,
		Language:        "vi", // Default, can be detected from locale
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Store provider data as JSON
	providerData := map[string]interface{}{
		"picture": userInfo.Picture,
		"locale":  userInfo.Locale,
	}
	dataJSON, _ := json.Marshal(providerData)
	dataStr := string(dataJSON)
	newUser.AuthProviderData = &dataStr

	// Save user
	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to create user from OAuth")
	}

	// Generate tokens
	return s.generateTokens(ctx, newUser)
}

// generateTokens creates access and refresh tokens for a user
func (s *Service) generateTokens(ctx context.Context, user *users.User) (*appAuth.LoginResult, error) {
	// Get couple ID if exists (not implemented here, can be added later)
	var coupleID *uuid.UUID
	// TODO: Fetch couple ID from couple service/repo if needed

	// Generate access token
	accessToken, err := s.jwtManager.GenerateAccessToken(
		user.ID,
		user.Username,
		user.Email,
		coupleID,
	)
	if err != nil {
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to generate access token")
	}

	// Generate refresh token ID
	refreshTokenID, err := s.jwtManager.GenerateRefreshTokenID()
	if err != nil {
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to generate refresh token ID")
	}

	// Store refresh token hash
	refreshToken := &domainAuth.RefreshToken{
		ID:          uuid.New(),
		UserID:      user.ID,
		TokenHash:   refreshTokenID, // TODO: hash this before storing (same as in auth service)
		ExpiresAt:   time.Now().Add(s.refreshTTL),
		DeviceInfo:  stringPtr("oauth"),
		IsRevoked:   false,
		CreatedAt:   time.Now(),
	}

	if err := s.tokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to store refresh token")
	}

	return &appAuth.LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenID,
		ExpiresIn:    int64(s.accessTTL.Seconds()),
		User: &users.User{
			ID:            user.ID,
			Username:      user.Username,
			Email:         user.Email,
			FirstName:     user.FirstName,
			LastName:      user.LastName,
			AvatarURL:     user.AvatarURL,
			IsActive:      user.IsActive,
			EmailVerified: user.EmailVerified,
		},
	}, nil
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func generateUsername(userInfo *GoogleUserInfo) string {
	// Generate a username from Google name
	base := userInfo.FirstName + userInfo.LastName
	if base == "" {
		// Fallback to email prefix
		atIdx := 0
		for i, c := range userInfo.Email {
			if c == '@' {
				atIdx = i
				break
			}
		}
		if atIdx > 0 {
			base = userInfo.Email[:atIdx]
		} else {
			base = "user"
		}
	}
	return base
}
