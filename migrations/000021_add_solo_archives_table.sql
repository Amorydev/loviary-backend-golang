-- +goose Up
-- +goose StatementBegin

CREATE TABLE solo_archives (
  archive_id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id                 UUID        NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  original_couple_id      UUID,
  subscription_expires_at TIMESTAMPTZ,
  is_active               BOOLEAN     DEFAULT TRUE,
  created_at              TIMESTAMPTZ DEFAULT NOW()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS solo_archives CASCADE;

-- +goose StatementEnd
