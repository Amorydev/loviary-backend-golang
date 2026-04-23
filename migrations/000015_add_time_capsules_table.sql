-- +goose Up
-- +goose StatementBegin

CREATE TABLE time_capsules (
  capsule_id      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  couple_id       UUID        NOT NULL REFERENCES couples(couple_id) ON DELETE CASCADE,
  created_by      UUID        REFERENCES users(user_id),
  title           VARCHAR(200),
  message         TEXT,
  voice_url       TEXT,
  image_urls      TEXT[],
  unlock_date     DATE        NOT NULL,
  is_sealed       BOOLEAN     DEFAULT FALSE,
  user1_confirmed BOOLEAN     DEFAULT FALSE,
  user2_confirmed BOOLEAN     DEFAULT FALSE,
  is_opened       BOOLEAN     DEFAULT FALSE,
  created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS time_capsules CASCADE;

-- +goose StatementEnd
