package dto

// DashboardResponse is the full dashboard payload returned by GET /home.
type DashboardResponse struct {
	User        UserInfo       `json:"user"`
	Couple      *CoupleInfo    `json:"couple,omitempty"`
	Chapter     *ChapterInfo   `json:"chapter,omitempty"`
	TodaysMood  TodaysMoodInfo `json:"todays_mood"`
	Streaks     []StreakInfo   `json:"streaks"`
	DailySpark  *SparkInfo     `json:"daily_spark,omitempty"`
	LastUpdated string         `json:"last_updated"`
}

// UserInfo contains the current user's display data.
type UserInfo struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

// CoupleInfo contains couple + partner display data.
type CoupleInfo struct {
	CoupleID         string `json:"couple_id"`
	PartnerName      string `json:"partner_name"`
	PartnerAvatarURL string `json:"partner_avatar_url"`
	RelationshipType string `json:"relationship_type"`
}

// ChapterInfo contains the current chapter progress data.
type ChapterInfo struct {
	Title            string `json:"title"`
	DaysTogether     int    `json:"days_together"`
	MilestoneTarget  int    `json:"milestone_target"`
	MilestonePercent int    `json:"milestone_percent"`
	StartDate        string `json:"start_date"`
	CoverImageURL    string `json:"cover_image_url"`
}

// MoodEntry holds a single mood entry (my_mood or partner_mood).
type MoodEntry struct {
	MoodType  string `json:"mood_type"`
	Intensity int    `json:"intensity"`
	Icon      string `json:"icon"`
}

// TodaysMoodInfo holds both user and partner mood for today.
type TodaysMoodInfo struct {
	MyMood      *MoodEntry `json:"my_mood"`
	PartnerMood *MoodEntry `json:"partner_mood"`
}

// DayLog represents a single day in the weekly_log.
type DayLog struct {
	Day       string `json:"day"`       // "Mon", "Tue", ...
	Completed bool   `json:"completed"` // true = both users logged
}

// StreakInfo contains streak data with weekly activity log.
type StreakInfo struct {
	ActivityType  string   `json:"activity_type"`
	CurrentStreak int      `json:"current_streak"`
	LongestStreak int      `json:"longest_streak"`
	Status        string   `json:"status"`
	WeeklyLog     []DayLog `json:"weekly_log"`
}

// SparkInfo contains today's daily spark question.
type SparkInfo struct {
	SparkID    string `json:"spark_id"`
	Question   string `json:"question"`
	Category   string `json:"category"`
	IsAnswered bool   `json:"is_answered"`
}
