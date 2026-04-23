-- +goose Up
-- +goose StatementBegin

CREATE TABLE memories (
  memory_id     UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  album_id      UUID        REFERENCES memory_albums(album_id) ON DELETE SET NULL,
  couple_id     UUID        NOT NULL REFERENCES couples(couple_id) ON DELETE CASCADE,
  created_by    UUID        REFERENCES users(user_id),
  title         VARCHAR(200),
  description   TEXT,
  memory_date   DATE        NOT NULL,
  location_name VARCHAR(200),
  location_lat  DECIMAL(10,8),
  location_lng  DECIMAL(11,8),
  category      VARCHAR(50),
  is_favorite   BOOLEAN     DEFAULT FALSE,
  is_hidden     BOOLEAN     DEFAULT FALSE,
  created_at    TIMESTAMPTZ DEFAULT NOW(),
  updated_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_memories_couple_date ON memories(couple_id, memory_date DESC);
CREATE INDEX idx_memories_album       ON memories(album_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS memories CASCADE;

-- +goose StatementEnd
