-- +goose Up
-- +goose StatementBegin

CREATE TABLE avatars (
  avatar_id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id            UUID        NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  couple_id          UUID        REFERENCES couples(couple_id),
  avatar_type        VARCHAR(10) DEFAULT 'personal'
                       CHECK (avatar_type IN ('personal','couple')),
  level              SMALLINT    DEFAULT 1,
  current_outfit_key VARCHAR(100),
  is_active          BOOLEAN     DEFAULT TRUE,
  updated_at         TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(user_id, avatar_type)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS avatars CASCADE;

-- +goose StatementEnd
