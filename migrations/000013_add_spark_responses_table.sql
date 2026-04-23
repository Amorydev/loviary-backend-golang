-- +goose Up
-- +goose StatementBegin

CREATE TABLE spark_responses (
  response_id  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  spark_id     UUID        NOT NULL REFERENCES daily_sparks(spark_id),
  user_id      UUID        NOT NULL REFERENCES users(user_id),
  couple_id    UUID        NOT NULL REFERENCES couples(couple_id) ON DELETE CASCADE,
  answer       TEXT        NOT NULL CHECK (char_length(answer) <= 500),
  responded_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(spark_id, user_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS spark_responses CASCADE;

-- +goose StatementEnd
