-- +goose Up
-- +goose StatementBegin

CREATE TABLE memory_files (
  file_id     UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  memory_id   UUID        NOT NULL REFERENCES memories(memory_id) ON DELETE CASCADE,
  file_type   VARCHAR(10) NOT NULL CHECK (file_type IN ('image','video','audio')),
  file_url    TEXT        NOT NULL,
  file_name   VARCHAR(255),
  file_size   INTEGER     CHECK (file_size > 0),
  duration    SMALLINT,
  width       SMALLINT,
  height      SMALLINT,
  caption     TEXT,
  sort_order  SMALLINT    DEFAULT 0,
  uploaded_by UUID        REFERENCES users(user_id),
  created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_memory_files_memory ON memory_files(memory_id, sort_order);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS memory_files CASCADE;

-- +goose StatementEnd
