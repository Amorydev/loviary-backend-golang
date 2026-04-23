-- +goose Up
-- +goose StatementBegin

CREATE TABLE love_goals (
  goal_id       UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  couple_id     UUID        NOT NULL REFERENCES couples(couple_id) ON DELETE CASCADE,
  created_by    UUID        NOT NULL REFERENCES users(user_id),
  title         VARCHAR(200) NOT NULL,
  category      VARCHAR(20) CHECK (category IN ('finance','travel','health','experience')),
  target_value  DECIMAL(12,2),
  current_value DECIMAL(12,2) DEFAULT 0,
  unit          VARCHAR(50),
  is_completed  BOOLEAN     DEFAULT FALSE,
  completed_at  TIMESTAMPTZ,
  created_at    TIMESTAMPTZ DEFAULT NOW(),
  updated_at    TIMESTAMPTZ DEFAULT NOW()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS love_goals CASCADE;

-- +goose StatementEnd
