-- +goose Up
-- +goose StatementBegin

-- Bỏ các cột definition (đã chuyển sang chapter_definitions)
ALTER TABLE love_chapters
  DROP COLUMN IF EXISTS title,
  DROP COLUMN IF EXISTS day_start,
  DROP COLUMN IF EXISTS day_end,
  DROP COLUMN IF EXISTS cover_art_url;

-- FK tới chapter_definitions
ALTER TABLE love_chapters
  ADD CONSTRAINT fk_love_chapters_chapter_def
  FOREIGN KEY (chapter_number) REFERENCES chapter_definitions(chapter_number);

-- Track ai trigger unlock
ALTER TABLE love_chapters
  ADD COLUMN IF NOT EXISTS unlocked_by UUID REFERENCES users(user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE love_chapters DROP CONSTRAINT IF EXISTS fk_love_chapters_chapter_def;
ALTER TABLE love_chapters DROP COLUMN IF EXISTS unlocked_by;
ALTER TABLE love_chapters ADD COLUMN IF NOT EXISTS title VARCHAR(100);
ALTER TABLE love_chapters ADD COLUMN IF NOT EXISTS day_start INTEGER;
ALTER TABLE love_chapters ADD COLUMN IF NOT EXISTS day_end INTEGER;
ALTER TABLE love_chapters ADD COLUMN IF NOT EXISTS cover_art_url TEXT;

-- +goose StatementEnd
