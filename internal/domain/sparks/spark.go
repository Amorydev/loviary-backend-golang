package sparks

import (
	"time"

	"github.com/google/uuid"
)

// DailySpark represents the spark assigned to a couple for a given day.
type DailySpark struct {
	SparkID    uuid.UUID `db:"spark_id"      json:"spark_id"`
	Question   string    `db:"question_text" json:"question"`
	Category   string    `db:"category"      json:"category"`
	IsAnswered bool      `db:"-"             json:"is_answered"` // derived: check spark_responses
}

// SparkAssignment represents the assignment of a spark to a couple for a date.
type SparkAssignment struct {
	AssignmentID uuid.UUID `db:"assignment_id" json:"assignment_id"`
	SparkID      uuid.UUID `db:"spark_id"      json:"spark_id"`
	CoupleID     uuid.UUID `db:"couple_id"     json:"couple_id"`
	AssignedDate time.Time `db:"assigned_date" json:"assigned_date"`
	ExpiresAt    time.Time `db:"expires_at"    json:"expires_at"`
}

// SparkResponse represents a user's response to a daily spark.
type SparkResponse struct {
	ResponseID  uuid.UUID `db:"response_id"  json:"response_id"`
	SparkID     uuid.UUID `db:"spark_id"     json:"spark_id"`
	UserID      uuid.UUID `db:"user_id"      json:"user_id"`
	CoupleID    uuid.UUID `db:"couple_id"    json:"couple_id"`
	Answer      string    `db:"answer"       json:"answer"`
	RespondedAt time.Time `db:"responded_at" json:"responded_at"`
}
