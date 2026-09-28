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

**[Read the documentation](docs/README.md)** for configuration, authentication, architecture, rate limiting, storage/caching, and tests.

## License

MIT — see [LICENSE](LICENSE).
