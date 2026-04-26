-- +goose Up
-- Add social authentication fields to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS auth_provider VARCHAR(20) DEFAULT 'password';
ALTER TABLE users ADD COLUMN IF NOT EXISTS auth_provider_id VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS auth_provider_data JSONB;

-- Create unique index for provider-based lookups
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_auth_provider_id ON users(auth_provider, auth_provider_id) WHERE auth_provider IS NOT NULL AND auth_provider_id IS NOT NULL;

-- +goose Down
-- Remove social authentication fields
DROP INDEX IF EXISTS idx_users_auth_provider_id;
ALTER TABLE users DROP COLUMN IF EXISTS auth_provider;
ALTER TABLE users DROP COLUMN IF EXISTS auth_provider_id;
ALTER TABLE users DROP COLUMN IF EXISTS auth_provider_data;
