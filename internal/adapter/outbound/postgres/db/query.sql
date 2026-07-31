-- name: GetUserByID :one
SELECT users.id, users.username, users.email, users.is_verified, users.updated_at, users.created_at
FROM users
WHERE users.id = $1;

-- name: GetUserByEmail :one
SELECT users.id, users.username, users.email, users.is_verified, users.password, users.updated_at, users.created_at
FROM users
WHERE users.email = $1;

-- name: GetUserByEmailWithoutPassword :one
SELECT users.id, users.username, users.email, users.is_verified, users.updated_at, users.created_at
FROM users
WHERE users.email = $1;

-- name: CreateUser :one
INSERT INTO users (username, email, password)
VALUES ($1, $2, $3)
RETURNING users.id;

-- name: GetAllUsers :many
SELECT users.id, users.username, users.email, users.is_verified, users.updated_at, users.created_at
FROM users
ORDER BY users.id;

-- name: UpdateUser :exec
UPDATE users
SET username = $1, email = $2
WHERE users.id = $3;

-- name: DeleteUser :exec
DELETE FROM users
WHERE users.id = $1;

-- name: UpdateUserPassword :exec
UPDATE users
SET password = $1
WHERE users.id = $2;

-- name: CreateVerifyEmail :exec
INSERT INTO verify_emails (id, user_id, email, expired_at)
VALUES ($1, $2, $3, $4);

-- name: GetVerifyEmailByID :one
SELECT verify_emails.id, verify_emails.email, verify_emails.expired_at
FROM verify_emails
WHERE verify_emails.id = $1;

-- name: VerifyEmailByToken :exec
UPDATE users
SET is_verified = true
WHERE users.email = (SELECT verify_emails.email FROM verify_emails WHERE verify_emails.id = $1);

-- name: VerifyEmailDeleteToken :exec
DELETE FROM verify_emails
WHERE verify_emails.id = $1;

-- name: CreateForgotPasswordEmail :exec
INSERT INTO forgot_password_emails (id, user_id, email, expired_at)
VALUES ($1, $2, $3, $4);

-- name: GetForgotPasswordByID :one
SELECT forgot_password_emails.id, forgot_password_emails.user_id, forgot_password_emails.email, forgot_password_emails.expired_at
FROM forgot_password_emails
WHERE forgot_password_emails.id = $1;

-- name: DeleteForgotPasswordByID :exec
DELETE FROM forgot_password_emails
WHERE forgot_password_emails.id = $1;

-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (id, user_id, token_hash, family_id, parent_id, expired_at, created_at, ip_address, user_agent)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: GetRefreshTokenByHash :one
SELECT id, user_id, token_hash, family_id, parent_id, expired_at, used_at, revoked_at, created_at, ip_address, user_agent
FROM refresh_tokens
WHERE token_hash = $1;

-- name: UseRefreshToken :exec
UPDATE refresh_tokens
SET used_at = NOW()
WHERE id = $1 AND used_at IS NULL;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE id = $1;

-- name: RevokeRefreshTokenFamily :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE family_id = $1 AND revoked_at IS NULL;
