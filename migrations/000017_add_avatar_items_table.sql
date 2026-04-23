-- +goose Up
-- +goose StatementBegin

CREATE TABLE avatar_items (
  item_id     UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  avatar_id   UUID        NOT NULL REFERENCES avatars(avatar_id) ON DELETE CASCADE,
  item_type   VARCHAR(20) NOT NULL
                CHECK (item_type IN ('outfit','accessory','pet','background')),
  item_key    VARCHAR(100) NOT NULL,
  unlocked_at TIMESTAMPTZ DEFAULT NOW(),
  source      VARCHAR(20) CHECK (source IN ('streak_reward','iap','milestone')),
  UNIQUE(avatar_id, item_key)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS avatar_items CASCADE;

-- +goose StatementEnd
