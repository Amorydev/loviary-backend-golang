-- +goose Up
-- +goose StatementBegin

CREATE TABLE mood_responses (
  response_id        UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  original_mood_id   UUID        NOT NULL REFERENCES moods(mood_id) ON DELETE CASCADE,
  responding_user_id UUID        NOT NULL REFERENCES users(user_id),
  response_text      TEXT        CHECK (char_length(response_text) <= 200),
  response_emoji     VARCHAR(10),
  created_at         TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(original_mood_id, responding_user_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS mood_responses CASCADE;

-- +goose StatementEnd
