package home

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	appChapters "loviary.app/backend/internal/application/chapters"
	appSparks "loviary.app/backend/internal/application/sparks"
	"loviary.app/backend/internal/application/couples"
	"loviary.app/backend/internal/application/moods"
	"loviary.app/backend/internal/application/streaks"
	"loviary.app/backend/internal/application/users"
	domainMoods "loviary.app/backend/internal/domain/moods"
	domainSparks "loviary.app/backend/internal/domain/sparks"
	domainStreaks "loviary.app/backend/internal/domain/streaks"
	domainUsers "loviary.app/backend/internal/domain/users"
	apperrors "loviary.app/backend/pkg/errors"
)

// Service aggregates data for the home dashboard.
type Service struct {
	userService    *users.Service
	coupleService  *couples.Service
	moodService    *moods.Service
	streakService  *streaks.Service
	chapterService *appChapters.Service
	sparkService   *appSparks.Service
}

// NewService creates a new home service.
func NewService(
	userService *users.Service,
	coupleService *couples.Service,
	moodService *moods.Service,
	streakService *streaks.Service,
	chapterService *appChapters.Service,
	sparkService *appSparks.Service,
) *Service {
	return &Service{
		userService:    userService,
		coupleService:  coupleService,
		moodService:    moodService,
		streakService:  streakService,
		chapterService: chapterService,
		sparkService:   sparkService,
	}
}

// --- Internal summary types ---

// UserSummary is kept for backward compatibility but DashboardData now carries
// the full *domainUsers.User so the HTTP layer can map it through dto.UserToResponse.

// CoupleSummary holds couple + partner info.
type CoupleSummary struct {
	CoupleID         uuid.UUID
	PartnerName      string
	PartnerAvatarURL string
	RelationshipType string
}

// ChapterSummary holds current chapter display info.
type ChapterSummary struct {
	Title            string
	DaysTogether     int
	MilestoneTarget  int
	MilestonePercent int
	StartDate        string
	CoverImageURL    string
}

// MoodSummary holds a single mood entry for the dashboard.
type MoodSummary struct {
	MoodType  string
	Intensity int
	Icon      string
}

// TodaysMoodSummary holds both moods for today.
type TodaysMoodSummary struct {
	MyMood      *MoodSummary
	PartnerMood *MoodSummary
}

// StreakSummary holds streak info with weekly log.
type StreakSummary struct {
	ActivityType  string
	CurrentStreak int
	LongestStreak int
	Status        string
	WeeklyLog     []domainStreaks.DayLog
}

// SparkSummary holds daily spark info.
type SparkSummary struct {
	SparkID    uuid.UUID
	Question   string
	Category   string
	IsAnswered bool
}

// DashboardData is the aggregated dashboard payload.
type DashboardData struct {
	User        *domainUsers.User
	Couple      *CoupleSummary
	Chapter     *ChapterSummary
	TodaysMood  TodaysMoodSummary
	Streaks     []StreakSummary
	DailySpark  *SparkSummary
	LastUpdated time.Time
	HasCouple   bool
}

// GetDashboard retrieves all data needed for the home dashboard in a single call.
func (s *Service) GetDashboard(ctx context.Context, userID uuid.UUID) (*DashboardData, error) {
	data := &DashboardData{
		LastUpdated: time.Now(),
	}

	// ── 1. Current user ──────────────────────────────────────────────────────
	user, err := s.userService.GetByID(ctx, userID)
	if err != nil {
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to get user")
	}
	data.User = user

	// ── 2. Active couple ─────────────────────────────────────────────────────
	couple, err := s.coupleService.GetActiveByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperrors.NoActiveCouple) || errors.Is(err, apperrors.ErrNotFound) {
			// No couple — return partial dashboard
			return data, nil
		}
		return nil, apperrors.New("INTERNAL_ERROR", "Failed to get couple")
	}
	data.HasCouple = true

	// ── 3. Partner info ──────────────────────────────────────────────────────
	partnerID, ok := couple.GetPartnerID(userID)
	if ok {
		partner, err := s.userService.GetByID(ctx, partnerID)
		if err == nil {
			data.Couple = &CoupleSummary{
				CoupleID:         couple.CoupleID,
				PartnerName:      partner.DisplayName(),
				PartnerAvatarURL: avatarURL(partner),
				RelationshipType: string(couple.RelationshipType),
			}
		}
	}
	if data.Couple == nil {
		data.Couple = &CoupleSummary{
			CoupleID:         couple.CoupleID,
			RelationshipType: string(couple.RelationshipType),
		}
	}

	// ── 4. Chapter ───────────────────────────────────────────────────────────
	if couple.RelationshipStartDate != nil {
		daysTogether := appChapters.CalculateDaysTogether(*couple.RelationshipStartDate)
		chapterView, err := s.chapterService.GetCurrentChapter(ctx, couple.CoupleID, daysTogether)
		if err == nil {
			data.Chapter = &ChapterSummary{
				Title:            chapterView.Definition.Title,
				DaysTogether:     daysTogether,
				MilestoneTarget:  chapterView.Definition.GetMilestoneTarget(),
				MilestonePercent: chapterView.MilestonePercent,
				StartDate:        couple.RelationshipStartDate.Format("2006-01-02"),
				CoverImageURL:    chapterView.Definition.GetCoverURL(),
			}
		}
	}

	// ── 5. Today's mood (my mood + partner mood) ─────────────────────────────
	myMood, err := s.moodService.GetTodaysMood(ctx, userID)
	if err == nil && myMood != nil {
		data.TodaysMood.MyMood = toMoodSummary(myMood)
	}
	if ok {
		partnerMood, err := s.moodService.GetTodaysMood(ctx, partnerID)
		if err == nil && partnerMood != nil && !partnerMood.IsShared {
			data.TodaysMood.PartnerMood = toMoodSummary(partnerMood)
		}
		// partner_mood = nil if not logged today or is_private
		if partnerMood != nil && partnerMood.IsShared {
			data.TodaysMood.PartnerMood = nil
		}
	}

	// ── 6. Streaks ───────────────────────────────────────────────────────────
	streakList, err := s.streakService.GetAllStreaks(ctx, couple.CoupleID)
	if err == nil {
		for _, st := range streakList {
			data.Streaks = append(data.Streaks, StreakSummary{
				ActivityType:  st.ActivityType,
				CurrentStreak: st.CurrentStreak,
				LongestStreak: st.LongestStreak,
				Status:        string(st.GetStreakStatus()),
				WeeklyLog:     st.WeeklyLog,
			})
		}
	}

	// ── 7. Daily spark ───────────────────────────────────────────────────────
	spark, err := s.sparkService.GetTodaySpark(ctx, couple.CoupleID, userID)
	if err == nil && spark != nil {
		data.DailySpark = toSparkSummary(spark)
	}

	return data, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func avatarURL(u *domainUsers.User) string {
	if u.AvatarURL != nil {
		return *u.AvatarURL
	}
	return ""
}

func toMoodSummary(m *domainMoods.Mood) *MoodSummary {
	icon := ""
	if m.MoodEmoji != nil {
		icon = *m.MoodEmoji
	} else {
		icon = moodDefaultIcon(string(m.MoodType))
	}
	return &MoodSummary{
		MoodType:  string(m.MoodType),
		Intensity: m.Intensity,
		Icon:      icon,
	}
}

// moodDefaultIcon returns a default emoji when mood_emoji is not set.
func moodDefaultIcon(moodType string) string {
	icons := map[string]string{
		"happy":     "😊",
		"sad":       "😢",
		"angry":     "😠",
		"anxious":   "😰",
		"calm":      "😌",
		"excited":   "🥳",
		"tired":     "😴",
		"love":      "🥰",
		"stressed":  "😤",
		"grateful":  "🙏",
		"neutral":   "😐",
	}
	if icon, ok := icons[moodType]; ok {
		return icon
	}
	return "😶"
}

func toSparkSummary(spark *domainSparks.DailySpark) *SparkSummary {
	return &SparkSummary{
		SparkID:    spark.SparkID,
		Question:   spark.Question,
		Category:   spark.Category,
		IsAnswered: spark.IsAnswered,
	}
}
