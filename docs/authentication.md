# Authentication and API

## Token lifecycle

1. Registration stores an Argon2id password hash and sends a verification email.
2. Successful password login sets a short-lived JWT access token and a random opaque refresh token in HttpOnly, SameSite=Lax cookies (`Secure` in production).
3. The access token is signed, expires after 15 minutes, and is verified without a database lookup. Refresh tokens contain no JWT claims; only their HMAC-SHA-256 digests are stored in PostgreSQL.
4. Refresh rotates the one-time token. Reusing an already-used token revokes its entire token family. Logout revokes the family and clears the cookies.
5. Password recovery and verification tokens are stored in PostgreSQL; [rate limits](rate-limiting.md) protect these workflows.

```mermaid
sequenceDiagram
    participant Client
    participant HTTP as HTTP adapter
    participant Auth as Auth service
    participant Tokens as JWT/token adapter
    participant DB as PostgreSQL
    Client->>HTTP: Login with credentials
    HTTP->>Auth: Login
    Auth->>DB: Load user and password hash
    Auth->>Tokens: Sign access JWT; generate opaque refresh token and HMAC
    Auth->>DB: Store refresh-token HMAC and family ID
    Auth-->>HTTP: Access and refresh tokens
    HTTP-->>Client: Set HttpOnly cookies
    Client->>HTTP: Refresh with refresh cookie
    HTTP->>Auth: RefreshTokens
    Auth->>Tokens: HMAC the presented token
    Auth->>Tokens: Sign new JWT; generate replacement refresh token and HMAC
    Auth->>DB: Consume old token and create replacement in a transaction
    Auth-->>HTTP: New access and refresh tokens
    HTTP-->>Client: Replace cookies
    Client->>HTTP: Reuse old refresh token
    HTTP->>Auth: RefreshTokens
    Auth->>DB: Revoke token family in a transaction
    HTTP-->>Client: 401 unauthorized
```

Existing JWT refresh sessions are invalidated by the security migration. Legacy bcrypt password hashes are not accepted; affected users must reset their password.

## Routes

| Area | Endpoints |
| --- | --- |
| Authentication | `POST /api/auth/register`, `/login`, `/refresh`, `/logout`, `/forgot-password`, `/reset-password`; `GET /api/auth/verify/email`; `POST /api/auth/verify/email/resend` |
| OAuth | `GET /api/oauth/:provider/`, `/api/oauth/:provider/callback` (`github` and `google`) |
| Users | `GET /api/user/me`, `/api/user/:id`, `/api/user/get-all`, `/api/user/email`; `PATCH /api/user/update`; `DELETE /api/user/delete/:id` |

See the generated [Swagger spec](swagger.yaml) or the development-only Swagger UI for request and response details. Handler routing lives in `internal/adapter/inbound/http/routes.go`.
