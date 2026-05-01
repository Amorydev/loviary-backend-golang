-- +goose Up
-- +goose StatementBegin

DROP TABLE IF EXISTS streaks CASCADE;

-- Streak là per-couple: cả 2 phải log thì current_streak mới tăng
-- Activity type duy nhất Sprint 1: mood_log
CREATE TABLE streaks (
  streak_id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  couple_id           UUID        NOT NULL REFERENCES couples(couple_id) ON DELETE CASCADE,
  activity_type       VARCHAR(50) NOT NULL,
  current_streak      INTEGER     NOT NULL DEFAULT 0,
  longest_streak      INTEGER     NOT NULL DEFAULT 0,
  last_completed_date DATE,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(couple_id, activity_type)
);

CREATE INDEX idx_streaks_couple ON streaks(couple_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS streaks CASCADE;

-- +goose StatementEnd
