CREATE TABLE IF NOT EXISTS rate_limit_buckets (
    key_hash TEXT PRIMARY KEY,
    count BIGINT NOT NULL,
    window_start TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_rate_limit_buckets_expires_at
    ON rate_limit_buckets(expires_at);
