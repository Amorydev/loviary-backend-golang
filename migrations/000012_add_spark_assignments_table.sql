-- +goose Up
-- +goose StatementBegin

CREATE TABLE spark_assignments (
  assignment_id UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  spark_id      UUID        NOT NULL REFERENCES daily_sparks(spark_id),
  couple_id     UUID        NOT NULL REFERENCES couples(couple_id) ON DELETE CASCADE,
  assigned_date DATE        NOT NULL,
  expires_at    TIMESTAMPTZ NOT NULL,
  UNIQUE(couple_id, assigned_date)
);

CREATE INDEX idx_spark_assign_couple ON spark_assignments(couple_id, assigned_date DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS spark_assignments CASCADE;

-- +goose StatementEnd
