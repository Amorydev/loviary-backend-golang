package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appHome "loviary.app/backend/internal/application/home"
	"loviary.app/backend/internal/interfaces/http/dto"
	"loviary.app/backend/internal/interfaces/http/middleware"
	apperrors "loviary.app/backend/pkg/errors"
)

// HomeHandler handles home/dashboard HTTP requests.
type HomeHandler struct {
	homeService *appHome.Service
}

// NewHomeHandler creates a new home handler.
func NewHomeHandler(homeService *appHome.Service) *HomeHandler {
	return &HomeHandler{homeService: homeService}
}

// GetDashboard retrieves aggregated dashboard data.
// @Summary      Get dashboard
// @Description  Aggregate endpoint — returns all data needed for the Home screen in one request
// @Tags         home
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.DashboardResponse
// @Failure      401  {object}  handlers.ErrorResponse
// @Failure      500  {object}  handlers.ErrorResponse
// @Router       /home [get]
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

	resp := mapDashboardToDTO(data)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

// mapDashboardToDTO converts the service DashboardData into the HTTP DTO.
func mapDashboardToDTO(data *appHome.DashboardData) dto.DashboardResponse {
	resp := dto.DashboardResponse{
		User:        dto.UserToResponse(data.User, data.HasCouple),
		TodaysMood:  mapTodaysMood(data.TodaysMood),
		Streaks:     mapStreaks(data.Streaks),
		LastUpdated: data.LastUpdated.Format("2006-01-02T15:04:05Z07:00"),
	}

	if data.Couple != nil {
		resp.Couple = &dto.CoupleInfo{
			CoupleID:         data.Couple.CoupleID.String(),
			PartnerName:      data.Couple.PartnerName,
			PartnerAvatarURL: data.Couple.PartnerAvatarURL,
			RelationshipType: data.Couple.RelationshipType,
		}
	}

	if data.Chapter != nil {
		resp.Chapter = &dto.ChapterInfo{
			Title:            data.Chapter.Title,
			DaysTogether:     data.Chapter.DaysTogether,
			MilestoneTarget:  data.Chapter.MilestoneTarget,
			MilestonePercent: data.Chapter.MilestonePercent,
			StartDate:        data.Chapter.StartDate,
			CoverImageURL:    data.Chapter.CoverImageURL,
		}
	}

	if data.DailySpark != nil {
		resp.DailySpark = &dto.SparkInfo{
			SparkID:    data.DailySpark.SparkID.String(),
			Question:   data.DailySpark.Question,
			Category:   data.DailySpark.Category,
			IsAnswered: data.DailySpark.IsAnswered,
		}
	}

	return resp
}

func mapTodaysMood(m appHome.TodaysMoodSummary) dto.TodaysMoodInfo {
	result := dto.TodaysMoodInfo{}
	if m.MyMood != nil {
		result.MyMood = &dto.MoodEntry{
			MoodType:  m.MyMood.MoodType,
			Intensity: m.MyMood.Intensity,
			Icon:      m.MyMood.Icon,
		}
	}
	if m.PartnerMood != nil {
		result.PartnerMood = &dto.MoodEntry{
			MoodType:  m.PartnerMood.MoodType,
			Intensity: m.PartnerMood.Intensity,
			Icon:      m.PartnerMood.Icon,
		}
	}
	return result
}

func mapStreaks(streaks []appHome.StreakSummary) []dto.StreakInfo {
	if len(streaks) == 0 {
		return []dto.StreakInfo{}
	}
	result := make([]dto.StreakInfo, len(streaks))
	for i, s := range streaks {
		weeklyLog := make([]dto.DayLog, len(s.WeeklyLog))
		for j, d := range s.WeeklyLog {
			weeklyLog[j] = dto.DayLog{Day: d.Day, Completed: d.Completed}
		}
		result[i] = dto.StreakInfo{
			ActivityType:  s.ActivityType,
			CurrentStreak: s.CurrentStreak,
			LongestStreak: s.LongestStreak,
			Status:        s.Status,
			WeeklyLog:     weeklyLog,
		}
	}
	return result
}
