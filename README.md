# auth-go

A production-minded Go authentication starter with Gin, PostgreSQL, OAuth, email verification, password reset, and a security-focused token lifecycle. Its hexagonal structure is designed to be extended without a large framework.

## Highlights

- Argon2id passwords and short-lived JWT access tokens.
- Opaque, rotating refresh tokens with replay detection and family-wide revocation.
- GitHub/Google OAuth, email verification, and password recovery.
- Configurable authentication rate limiting, audit logs, and optional Prometheus metrics.

## Quick start

Requires Go 1.25+, Docker Compose, and the `migrate` CLI.

```bash
git clone https://github.com/Jonathan0823/auth-go.git
cd auth-go
cp .env.example .env
# Set the secrets and email credentials in .env before running.
go mod download
docker compose up -d db redis
make migrate-up
make run
```

The API listens at `http://localhost:8080` by default. Swagger UI is available at `/swagger/index.html` when `ENABLE_SWAGGER=true` outside production.

## Documentation

- [Setup and operations](docs/setup.md) — configuration, migrations, testing, and observability.
- [Authentication and API](docs/authentication.md) — tokens, cookies, and routes.
- [Rate limiting](docs/rate-limiting.md) — policies, backends, and failure handling.
- [Storage and caching](docs/storage-and-caching.md) — PostgreSQL, Redis, and what is not cached.
- [Architecture](docs/architecture.md) — package boundaries and extension points.
- [Swagger specification](docs/swagger.yaml) — generated HTTP API contract.

## License

MIT — see [LICENSE](LICENSE).
