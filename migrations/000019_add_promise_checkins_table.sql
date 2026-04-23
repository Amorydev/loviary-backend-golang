-- +goose Up
-- +goose StatementBegin

CREATE TABLE promise_checkins (
  checkin_id   UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  promise_id   UUID        NOT NULL REFERENCES promises(promise_id) ON DELETE CASCADE,
  user_id      UUID        NOT NULL REFERENCES users(user_id),
  status       VARCHAR(5)  NOT NULL CHECK (status IN ('yes','no')),
  checkin_date DATE        NOT NULL,
  note         TEXT,
  created_at   TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(promise_id, user_id, checkin_date)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS promise_checkins CASCADE;

-- +goose StatementEnd
