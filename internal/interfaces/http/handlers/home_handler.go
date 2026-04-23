package handlers

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"

    appHome "loviary.app/backend/internal/application/home"
    "loviary.app/backend/internal/interfaces/http/middleware"
    apperrors "loviary.app/backend/pkg/errors"
)

// HomeHandler handles home/dashboard HTTP requests
type HomeHandler struct {
    homeService *appHome.Service
}

// NewHomeHandler creates a new home handler
func NewHomeHandler(homeService *appHome.Service) *HomeHandler {
    return &HomeHandler{homeService: homeService}
}

// DashboardResponse represents dashboard response
type DashboardResponse struct {
    UserID           uuid.UUID             `json:"user_id"`
    Couple           *struct {
        CoupleID         uuid.UUID  `json:"couple_id"`
        RelationshipType string     `json:"relationship_type"`
        PartnerName      *string    `json:"partner_name,omitempty"`
    } `json:"couple,omitempty"`
    TodaysMood       *struct {
        MoodType   string `json:"mood_type"`
        Intensity  int    `json:"intensity"`
    } `json:"todays_mood,omitempty"`
    MoodHistory      []struct {
        Date       string    `json:"date"`
        MoodType   string    `json:"mood_type"`
        Intensity  int       `json:"intensity"`
    } `json:"mood_history,omitempty"`
    Streaks          []struct {
        ActivityType   string `json:"activity_type"`
        CurrentStreak  int    `json:"current_streak"`
        LongestStreak  int    `json:"longest_streak"`
        Status         string `json:"status"`
    } `json:"streaks,omitempty"`
    LastUpdated      time.Time `json:"last_updated"`
}

// GetDashboard retrieves aggregated dashboard data
// @Summary Get dashboard
// @Description Get aggregated dashboard data including couple info, mood, streaks
// @Tags home
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success  200  {object}  handlers.DashboardResponse "Dashboard data"
// @Failure  401  {object}  handlers.ErrorResponse "Not authenticated"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /home [get]
func (h *HomeHandler) GetDashboard(c *gin.Context) {
    userID, exists := middleware.GetUserID(c)
    if !exists {
        c.JSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
        return
    }

    data, err := h.homeService.GetDashboard(c.Request.Context(), userID)
    if err != nil {
        c.Error(err)
        return
    }

    // Transform couple data
    var coupleResp *struct {
        CoupleID         uuid.UUID  `json:"couple_id"`
        RelationshipType string     `json:"relationship_type"`
        PartnerName      *string    `json:"partner_name,omitempty"`
    }
    if data.Couple != nil {
        coupleResp = &struct {
            CoupleID         uuid.UUID  `json:"couple_id"`
            RelationshipType string     `json:"relationship_type"`
            PartnerName      *string    `json:"partner_name,omitempty"`
        }{
            CoupleID:         data.Couple.CoupleID,
            RelationshipType: string(data.Couple.RelationshipType),
        }
    }

    // Transform mood history
    moodHistory := make([]struct {
        Date       string    `json:"date"`
        MoodType   string    `json:"mood_type"`
        Intensity  int       `json:"intensity"`
    }, len(data.MoodHistory))
    for i, mood := range data.MoodHistory {
        moodHistory[i] = struct {
            Date       string    `json:"date"`
            MoodType   string    `json:"mood_type"`
            Intensity  int       `json:"intensity"`
        }{
            Date:       mood.Date.Format("2006-01-02"),
            MoodType:   string(mood.MoodType),
            Intensity:  mood.Intensity,
        }
    }

    // Transform streaks
    streaks := make([]struct {
        ActivityType   string `json:"activity_type"`
        CurrentStreak  int    `json:"current_streak"`
        LongestStreak  int    `json:"longest_streak"`
        Status         string `json:"status"`
    }, len(data.Streaks))
    for i, streak := range data.Streaks {
        streaks[i] = struct {
            ActivityType   string `json:"activity_type"`
            CurrentStreak  int    `json:"current_streak"`
            LongestStreak  int    `json:"longest_streak"`
            Status         string `json:"status"`
        }{
            ActivityType:   streak.ActivityType,
            CurrentStreak:  streak.CurrentStreak,
            LongestStreak:  streak.LongestStreak,
            Status:         string(streak.GetStreakStatus()),
        }
    }

    resp := DashboardResponse{
        UserID:      data.UserID,
        Couple:      coupleResp,
        MoodHistory: moodHistory,
        Streaks:     streaks,
        LastUpdated: data.LastUpdated,
    }

    if data.TodaysMood != nil {
        resp.TodaysMood = &struct {
            MoodType  string `json:"mood_type"`
            Intensity int    `json:"intensity"`
        }{
            MoodType:  string(data.TodaysMood.MoodType),
            Intensity: data.TodaysMood.Intensity,
        }
    }

    c.JSON(http.StatusOK, resp)
}

// HomeSummaryResponse represents a simplified home summary
type HomeSummaryResponse struct {
    CoupleID         *uuid.UUID `json:"couple_id,omitempty"`
    PartnerName      *string    `json:"partner_name,omitempty"`
    RelationshipType *string    `json:"relationship_type,omitempty"`
    TodaysMoodType  *string    `json:"todays_mood_type,omitempty"`
    TodaysMoodIntensity *int   `json:"todays_mood_intensity,omitempty"`
    ActiveStreaks   int        `json:"active_streaks"`
    LongestStreak   int        `json:"longest_streak"`
}

// GetSummary retrieves a simplified home summary
// @Summary Get home summary
// @Description Get simplified summary for home widget display
// @Tags home
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success  200  {object}  handlers.HomeSummaryResponse "Summary data"
// @Failure  401  {object}  handlers.ErrorResponse "Not authenticated"
// @Failure  500  {object}  handlers.ErrorResponse "Internal server error"
// @Router   /home/summary [get]
func (h *HomeHandler) GetSummary(c *gin.Context) {
    userID, exists := middleware.GetUserID(c)
    if !exists {
        c.JSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
        return
    }

    summary, err := h.homeService.GetSummary(c.Request.Context(), userID)
    if err != nil {
        c.Error(err)
        return
    }

    resp := HomeSummaryResponse{
        ActiveStreaks:  summary.ActiveStreaks,
        LongestStreak:  summary.LongestStreak,
    }

    if summary.CoupleID != nil {
        resp.CoupleID = summary.CoupleID
        resp.RelationshipType = summary.RelationshipType
        resp.PartnerName = summary.PartnerName
    }

    if summary.TodaysMoodType != nil {
        resp.TodaysMoodType = summary.TodaysMoodType
        resp.TodaysMoodIntensity = summary.TodaysMoodIntensity
    }

    c.JSON(http.StatusOK, resp)
}
