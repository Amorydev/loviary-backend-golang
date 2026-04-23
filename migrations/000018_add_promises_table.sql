-- +goose Up
-- +goose StatementBegin

CREATE TABLE promises (
  promise_id      UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  couple_id       UUID        NOT NULL REFERENCES couples(couple_id) ON DELETE CASCADE,
  created_by      UUID        NOT NULL REFERENCES users(user_id),
  title           VARCHAR(200) NOT NULL,
  deadline        DATE,
  status          VARCHAR(20) DEFAULT 'PENDING'
                    CHECK (status IN ('PENDING','ACTIVE','COMPLETED','FAILED')),
  user1_confirmed BOOLEAN     DEFAULT FALSE,
  user2_confirmed BOOLEAN     DEFAULT FALSE,
  created_at      TIMESTAMPTZ DEFAULT NOW(),
  updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS promises CASCADE;

-- +goose StatementEnd
