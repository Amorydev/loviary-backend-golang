-- +goose Up
-- +goose StatementBegin

CREATE TABLE notifications (
  notification_id   UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id           UUID        NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  couple_id         UUID        REFERENCES couples(couple_id),
  title             VARCHAR(200) NOT NULL,
  message           TEXT,
  deep_link         TEXT,
  notification_type VARCHAR(30) NOT NULL
                      CHECK (notification_type IN (
                        'mood_sync','streak','reminder','spark',
                        'capsule','chapter','sos','pair_invite'
                      )),
  is_read           BOOLEAN     DEFAULT FALSE,
  scheduled_for     TIMESTAMPTZ,
  expire_at         TIMESTAMPTZ,
  created_at        TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_notif_user_unread ON notifications(user_id, created_at DESC)
  WHERE is_read = FALSE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS notifications CASCADE;

-- +goose StatementEnd
