-- +goose Up
-- +goose StatementBegin

CREATE TABLE reminders (
  reminder_id     UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  couple_id       UUID        NOT NULL REFERENCES couples(couple_id) ON DELETE CASCADE,
  reminder_type   VARCHAR(50) NOT NULL
                    CHECK (reminder_type IN (
                      'anniversary','birthday','custom','milestone','lunar'
                    )),
  title           VARCHAR(200) NOT NULL,
  description     TEXT,
  reminder_date   DATE,
  lunar_month_day VARCHAR(5),
  reminder_time   TIME        DEFAULT '09:00',
  is_recurring    BOOLEAN     DEFAULT TRUE,
  is_active       BOOLEAN     DEFAULT TRUE,
  next_occurrence TIMESTAMPTZ,
  created_by      UUID        REFERENCES users(user_id),
  created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_reminders_next ON reminders(couple_id, next_occurrence)
  WHERE is_active = TRUE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS reminders CASCADE;

-- +goose StatementEnd
