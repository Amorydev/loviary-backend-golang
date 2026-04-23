-- +goose Up
-- +goose StatementBegin

CREATE TABLE streaks (
  streak_id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  couple_id          UUID        NOT NULL REFERENCES couples(couple_id) ON DELETE CASCADE,
  current_count      INTEGER     DEFAULT 0,
  best_count         INTEGER     DEFAULT 0,
  last_streak_date   DATE,
  freeze_count       INTEGER     DEFAULT 0,
  total_freeze_used  INTEGER     DEFAULT 0,
  user1_logged_today BOOLEAN     DEFAULT FALSE,
  user2_logged_today BOOLEAN     DEFAULT FALSE,
  updated_at         TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(couple_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS streaks CASCADE;

-- +goose StatementEnd
