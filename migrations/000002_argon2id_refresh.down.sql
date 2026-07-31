DROP TABLE IF EXISTS refresh_tokens;

CREATE TABLE IF NOT EXISTS token_log (
    id UUID PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    jti VARCHAR(100) NOT NULL UNIQUE,
    refreshed_from_jti VARCHAR(100),
    invalidated_at TIMESTAMP,
    expired_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT NOT NULL,
    FOREIGN KEY (refreshed_from_jti) REFERENCES token_log(jti)
);

CREATE INDEX IF NOT EXISTS idx_token_log_user_id ON token_log(user_id);
CREATE INDEX IF NOT EXISTS idx_token_log_jti ON token_log(jti);

ALTER TABLE users ALTER COLUMN password TYPE VARCHAR(100);
