-- +goose Up
-- +goose StatementBegin

CREATE TABLE refresh_tokens (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID        NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  token_hash  TEXT        NOT NULL UNIQUE,
  expires_at  TIMESTAMPTZ NOT NULL,
  device_info VARCHAR(500),
  is_revoked  BOOLEAN     DEFAULT FALSE,
  created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id) WHERE is_revoked = FALSE;
CREATE INDEX idx_refresh_tokens_hash ON refresh_tokens(token_hash);

CREATE TABLE user_devices (
  token_id    UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID        NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  fcm_token   TEXT        NOT NULL,
  platform    VARCHAR(10) NOT NULL CHECK (platform IN ('ios','android','web')),
  device_name VARCHAR(100),
  is_active   BOOLEAN     DEFAULT TRUE,
  created_at  TIMESTAMPTZ DEFAULT NOW(),
  updated_at  TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(user_id, fcm_token)
);

CREATE INDEX idx_user_devices_user ON user_devices(user_id) WHERE is_active = TRUE;

CREATE TABLE couples (
  couple_id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user1_id                UUID        NOT NULL REFERENCES users(user_id),
  user2_id                UUID        REFERENCES users(user_id),
  couple_name             VARCHAR(100),
  relationship_start_date DATE,
  status                  VARCHAR(30) NOT NULL DEFAULT 'pending_invitation'
                            CHECK (status IN ('pending_invitation','active','grace_period','ended')),
  relationship_type       VARCHAR(20) DEFAULT 'dating'
                            CHECK (relationship_type IN ('dating','engaged','married')),
  invitation_expires_at   TIMESTAMPTZ,
  breakup_initiated_by    UUID        REFERENCES users(user_id),
  breakup_grace_until     TIMESTAMPTZ,
  created_at              TIMESTAMPTZ DEFAULT NOW(),
  updated_at              TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_couples_active_u1 ON couples(user1_id) WHERE status = 'active';
CREATE UNIQUE INDEX idx_couples_active_u2 ON couples(user2_id) WHERE status = 'active';
CREATE INDEX idx_couples_status ON couples(status);

CREATE TABLE subscriptions (
  subscription_id  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  couple_id        UUID        NOT NULL REFERENCES couples(couple_id) ON DELETE CASCADE,
  plan             VARCHAR(20) NOT NULL DEFAULT 'free'
                    CHECK (plan IN ('free','premium','premium_plus','lifetime','solo_archive')),
  started_at       TIMESTAMPTZ DEFAULT NOW(),
  expires_at       TIMESTAMPTZ,
  payment_provider VARCHAR(20) CHECK (payment_provider IN ('apple','google','stripe')),
  is_active        BOOLEAN     DEFAULT TRUE,
  UNIQUE(couple_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS subscriptions CASCADE;
DROP TABLE IF EXISTS couples CASCADE;
DROP TABLE IF EXISTS user_devices CASCADE;
DROP TABLE IF EXISTS refresh_tokens CASCADE;

-- +goose StatementEnd
