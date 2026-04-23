-- +goose Up
-- +goose StatementBegin

CREATE TABLE daily_sparks (
  spark_id        UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  question_text   TEXT        NOT NULL,
  category        VARCHAR(20) NOT NULL
                    CHECK (category IN ('fun','deep','memory','future')),
  is_ai_generated BOOLEAN     DEFAULT FALSE,
  created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS daily_sparks CASCADE;

-- +goose StatementEnd
