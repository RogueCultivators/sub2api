-- Add expires_at for group expiration configuration.
ALTER TABLE groups ADD COLUMN IF NOT EXISTS expires_at timestamptz;
COMMENT ON COLUMN groups.expires_at IS 'Group expiration time (NULL means no expiration).';
CREATE INDEX IF NOT EXISTS idx_groups_expires_at ON groups(expires_at);
