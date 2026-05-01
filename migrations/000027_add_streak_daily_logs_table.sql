-- +goose Up
-- +goose StatementBegin

-- Lưu log từng user từng ngày.
-- completed cho 1 ngày = cả 2 user trong couple đều có row cho ngày đó.
CREATE TABLE streak_daily_logs (
  log_id        UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  couple_id     UUID        NOT NULL REFERENCES couples(couple_id) ON DELETE CASCADE,
  user_id       UUID        NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  activity_type VARCHAR(50) NOT NULL,
  log_date      DATE        NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(couple_id, user_id, activity_type, log_date)
);

CREATE INDEX idx_streak_daily_logs_lookup
  ON streak_daily_logs(couple_id, activity_type, log_date DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS streak_daily_logs CASCADE;

-- +goose StatementEnd
