-- +goose Up
-- +goose StatementBegin

CREATE TABLE memory_albums (
  album_id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  couple_id         UUID        NOT NULL REFERENCES couples(couple_id) ON DELETE CASCADE,
  album_name        VARCHAR(100) NOT NULL,
  album_description TEXT,
  cover_image_url   TEXT,
  is_collaborative  BOOLEAN     DEFAULT TRUE,
  created_by        UUID        REFERENCES users(user_id),
  created_at        TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_albums_couple ON memory_albums(couple_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS memory_albums CASCADE;

-- +goose StatementEnd
