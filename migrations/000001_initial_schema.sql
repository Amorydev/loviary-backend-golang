-- +goose Up
-- +goose StatementBegin

CREATE TABLE users (
  user_id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username       VARCHAR(50)  UNIQUE NOT NULL,
  email          VARCHAR(100) UNIQUE NOT NULL,
  password_hash  VARCHAR(255) NOT NULL,
  first_name     VARCHAR(50),
  last_name      VARCHAR(50),
  date_of_birth  DATE,
  gender         VARCHAR(15) CHECK (gender IN ('male','female','other','prefer_not')),
  language       VARCHAR(10)  DEFAULT 'vi',
  key_couple     VARCHAR(20)  UNIQUE,
  avatar_url     TEXT,
  is_active      BOOLEAN      DEFAULT TRUE,
  email_verified BOOLEAN      DEFAULT FALSE,
  created_at     TIMESTAMPTZ  DEFAULT NOW(),
  updated_at     TIMESTAMPTZ  DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_key_couple ON users(key_couple);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS users CASCADE;

-- +goose StatementEnd
