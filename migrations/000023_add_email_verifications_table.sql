-- +goose Up
-- +goose StatementBegin

CREATE TABLE email_verifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    code            VARCHAR(6) NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    verified_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- Unique constraint: mỗi user chỉ có 1 verification record tại một thời điểm
CREATE UNIQUE INDEX IF NOT EXISTS idx_email_verifications_user_id ON email_verifications(user_id);

-- Index để lookup nhanh bằng code
CREATE INDEX IF NOT EXISTS idx_email_verifications_code ON email_verifications(code);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS email_verifications CASCADE;

-- +goose StatementEnd
