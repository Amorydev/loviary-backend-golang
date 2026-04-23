-- +goose Up
-- +goose StatementBegin

CREATE TABLE love_chapters (
  chapter_id     UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  couple_id      UUID        NOT NULL REFERENCES couples(couple_id) ON DELETE CASCADE,
  chapter_number SMALLINT    NOT NULL,
  title          VARCHAR(100) NOT NULL,
  day_start      INTEGER     NOT NULL,
  day_end        INTEGER,
  cover_art_url  TEXT,
  is_unlocked    BOOLEAN     DEFAULT FALSE,
  unlocked_at    TIMESTAMPTZ,
  UNIQUE(couple_id, chapter_number)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS love_chapters CASCADE;

-- +goose StatementEnd
