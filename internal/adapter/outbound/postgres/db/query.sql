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
INSERT INTO forgot_password_emails (id, email, expired_at)
VALUES ($1, $2, $3);

-- name: GetForgotPasswordByID :one
SELECT forgot_password_emails.id, forgot_password_emails.user_id, forgot_password_emails.email, forgot_password_emails.expired_at
FROM forgot_password_emails
WHERE forgot_password_emails.id = $1;

-- name: DeleteForgotPasswordByID :exec
DELETE FROM forgot_password_emails
WHERE forgot_password_emails.id = $1;

-- name: CreateTokenLog :exec
INSERT INTO token_log (id, user_id, jti, refreshed_from_jti, invalidated_at, expired_at, created_at, ip_address, user_agent)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: GetTokenLogByJTI :one
SELECT token_log.id, token_log.user_id, token_log.jti, token_log.refreshed_from_jti, token_log.invalidated_at, token_log.expired_at, token_log.created_at, token_log.ip_address, token_log.user_agent
FROM token_log
WHERE token_log.jti = $1;

-- name: InvalidateTokenLog :exec
UPDATE token_log
SET invalidated_at = NOW()
WHERE token_log.jti = $1;

-- name: InvalidateAndRefreshTokenLog :exec
UPDATE token_log
SET invalidated_at = NOW(), refreshed_from_jti = $2
WHERE token_log.jti = $1;

-- name: IsTokenLogInvalidated :one
SELECT EXISTS(SELECT 1 FROM token_log WHERE token_log.jti = $1 AND token_log.invalidated_at IS NOT NULL) AS invalidated;
