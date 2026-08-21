# auth-go

A production-minded authentication starter for Go teams who want secure defaults without a giant framework.

Build a REST API with Gin, PostgreSQL, OAuth, email verification, password reset, and a security-focused token lifecycle—ready to extend from a clean hexagonal architecture.

## Highlights

- **Argon2id passwords** — memory-hard password hashing with explicit parameters.
- **JWT access tokens** — short-lived tokens for authenticated API requests.
- **Opaque refresh tokens** — cryptographically random values with only HMAC-SHA-256 digests stored in PostgreSQL.
- **Refresh rotation and replay detection** — one-time refresh tokens with family-wide revocation.
- **Secure cookies** — HttpOnly, SameSite=Lax, and Secure in production.
- **OAuth login** — GitHub and Google providers.
- **Email workflows** — verification and password-reset flows.
- **Integration-ready CI** — PostgreSQL-backed tests run separately from the unit suite.
- **Bounded server lifecycle** — explicit HTTP timeouts and graceful SIGINT/SIGTERM shutdown.

## Quick start

### Prerequisites

- Go 1.25+
- Docker and Docker Compose
- PostgreSQL, or the included Compose database
- `migrate` CLI for integration tests and migrations

### Run locally

```bash
git clone https://github.com/Jonathan0823/auth-go.git
cd auth-go
cp .env.example .env

go mod download
docker compose up -d db redis
make migrate-up
make run
```

The API starts on `http://localhost:8080` by default.

With `ENABLE_SWAGGER=true` outside production, interactive API documentation is available at [`http://localhost:8080/swagger/index.html`](http://localhost:8080/swagger/index.html). The generated Swagger 2.0 document is served at `/swagger/doc.json`.

## Configuration

The application uses one PostgreSQL connection string everywhere: `DATABASE_URL`.

Start from `.env.example` and set at least:

```dotenv
DATABASE_URL=postgres://postgres:postgres@localhost:5432/auth_go?sslmode=disable
JWT_ACCESS_SECRET=replace-with-a-long-random-secret
REFRESH_TOKEN_HASH_KEY=replace-with-a-separate-long-random-secret
ENABLE_SWAGGER=true
ENABLE_METRICS=false
SESSION_SECRET=replace-with-a-long-random-secret
```

Email credentials are required because registration and recovery send mail. OAuth providers are optional, but each configured client ID must have its matching secret:

```dotenv
EMAIL=your-email@gmail.com
PASSWORD=your-gmail-app-password
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
```

Configuration is validated before the database or HTTP server starts. In production, JWT, refresh-token, session, and rate-limit secrets must each be at least 32 bytes.

Authentication rate limiting supports `memory`, `redis`, and `postgres` backends. Memory is intended for development or an explicitly approved single-instance deployment; production should use Redis or PostgreSQL. Set `RATE_LIMIT_KEY` to a separate secret and configure `TRUSTED_PROXIES` only for known proxy networks.

`make` loads `.env` automatically. Use another file explicitly when needed:

```bash
ENV_FILE=.env.prod make run
```

## Database and tests

```bash
make migrate-up                         # Apply migrations
make test                                # Unit tests
make integration                         # Migrate + PostgreSQL/Redis integration tests
make coverage                            # Integration tests + coverage/coverage.out
make vet                                 # Static checks
make sqlc                                # Regenerate PostgreSQL code
make swagger                             # Regenerate Swagger artifacts
make swagger-validate                    # Validate the Swagger document
```

Integration tests use the `integration` build tag and require PostgreSQL. Redis-backed tests run when `REDIS_ADDR` is configured:

```bash
go test -tags=integration ./...
```

## Observability

- `GET /health/live` reports process liveness without checking PostgreSQL.
- `GET /health/ready` reports database readiness with a bounded PostgreSQL ping.
- `GET /metrics` exposes Prometheus metrics only when `ENABLE_METRICS=true`; it is an operational endpoint and is intentionally excluded from Swagger.
- Security audit events are emitted as structured JSON logs with request IDs and safe categorical context.
- Rate-limit denials and backend failures use low-cardinality audit and Prometheus events without raw identifiers.

Metrics labels use route templates and avoid user-controlled values. Restrict `/metrics` to trusted monitoring systems in production. Grafana, Loki, and tracing infrastructure are intentionally not bundled.

## Authentication model

1. A successful login returns a short-lived JWT access token and an opaque refresh token in secure cookies.
2. Refresh tokens are random bearer values; no user data or JWT claims are embedded in them.
3. Only an HMAC digest of a refresh token is persisted.
4. Every refresh invalidates the previous token and issues a replacement.
5. Reusing a rotated token revokes the entire refresh-token family.
6. Access tokens remain stateless and expire after 15 minutes.

> **Migration note:** The security migration invalidates existing JWT refresh sessions. Legacy bcrypt password hashes are not accepted; affected users must reset their password.

## API documentation

Swagger documentation is generated from Go annotations and committed under `docs/`. To regenerate it after changing a handler contract:

```bash
make swagger
make swagger-validate
```

Swagger UI is opt-in and development-only. Set `ENABLE_SWAGGER=true` locally; it remains disabled whenever `ENVIRONMENT=production`.

## API overview

### Authentication

- `POST /api/auth/register`
- `POST /api/auth/login`
- `POST /api/auth/refresh`
- `POST /api/auth/logout`
- `POST /api/auth/forgot-password`
- `POST /api/auth/reset-password`
- `GET /api/auth/verify/email`
- `POST /api/auth/verify/email/resend`

### OAuth

- `GET /api/oauth/:provider/`
- `GET /api/oauth/:provider/callback`

Supported providers: GitHub and Google.

### Users

- `GET /api/user/me`
- `GET /api/user/:id`
- `GET /api/user/get-all`
- `GET /api/user/email`
- `PATCH /api/user/update`
- `DELETE /api/user/delete/:id`

## Project structure

```text
cmd/       application entrypoint
docs/      generated Swagger specification and registration
internal/
  core/    domain models, ports, and services
  adapter/ HTTP, PostgreSQL, JWT, password, OAuth, and email adapters
  platform/ configuration, database, logging, and server setup
migrations/ PostgreSQL schema migrations
```

## License

MIT — see [LICENSE](LICENSE).
