-- +goose Up
-- +goose StatementBegin

CREATE TABLE moods (
  mood_id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id          UUID        NOT NULL REFERENCES users(user_id),
  couple_id        UUID        NOT NULL REFERENCES couples(couple_id),
  mood_type        VARCHAR(20) NOT NULL
                     CHECK (mood_type IN ('happy','sad','angry','neutral','excited')),
  intensity        SMALLINT    CHECK (intensity BETWEEN 1 AND 10),
  mood_description TEXT        CHECK (char_length(mood_description) <= 300),
  mood_emoji       VARCHAR(10),
  color_tag        VARCHAR(7)  CHECK (color_tag ~ '^#[0-9A-Fa-f]{6}$'),
  voice_note_url   TEXT,
  voice_duration   SMALLINT,
  is_private       BOOLEAN     DEFAULT FALSE,
  weather_condition VARCHAR(50),
  logged_date      DATE        NOT NULL
                     DEFAULT (NOW() AT TIME ZONE 'Asia/Ho_Chi_Minh')::date,
  created_at       TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(user_id, logged_date)
);

CREATE INDEX idx_moods_couple_date ON moods(couple_id, logged_date DESC);
CREATE INDEX idx_moods_user_date   ON moods(user_id,   logged_date DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS moods CASCADE;

-- +goose StatementEnd
