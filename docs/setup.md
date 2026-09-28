# Setup and operations

## Prerequisites and configuration

Install Go 1.25+, Docker/Compose (or your own PostgreSQL), and the `migrate` CLI. Copy `.env.example` to `.env`; `make` loads `.env` automatically. To use another file, run `ENV_FILE=.env.prod make run`.

All database operations use `DATABASE_URL`. Before running the server, configure separate secrets and mail credentials:

```dotenv
DATABASE_URL=postgres://postgres:postgres@localhost:5432/auth_go?sslmode=disable
JWT_ACCESS_SECRET=replace-with-a-long-random-secret
REFRESH_TOKEN_HASH_KEY=replace-with-a-separate-long-random-secret
SESSION_SECRET=replace-with-a-third-long-random-secret
RATE_LIMIT_KEY=replace-with-a-fourth-long-random-secret
EMAIL=your-email@gmail.com
PASSWORD=your-gmail-app-password
```

GitHub and Google OAuth are optional (`GITHUB_CLIENT_ID`/`GITHUB_CLIENT_SECRET`, `GOOGLE_CLIENT_ID`/`GOOGLE_CLIENT_SECRET`); a configured client ID requires a matching secret. Configuration is validated before the database or HTTP server starts; production requires each secret above to be at least 32 bytes. Refer to `.env.example` for every supported setting, including `ALLOWED_ORIGINS`, `DOMAIN`, `LOG_LEVEL`, and [rate-limit policies](rate-limiting.md). Replace the example secrets and do not commit `.env`.

## Run and verify

```bash
docker compose up -d db redis  # Redis is optional unless RATE_LIMIT_BACKEND=redis
make migrate-up
make run
```

By default the API listens on `http://localhost:8080`. `GET /health/live` reports process liveness; `GET /health/ready` checks the database with a bounded ping.

```bash
make test               # Unit tests
make vet                # Static checks
make integration        # Run migrations and integration-tagged tests
make sqlc               # Regenerate PostgreSQL code
make swagger            # Regenerate Swagger from Go annotations
make swagger-validate   # Lint the generated Swagger document
```

**Integration tests write to the database in `DATABASE_URL`.** Use a dedicated, disposable test database (with migrations applied), never production or a shared application database. Redis-backed integration tests run only when `REDIS_ADDR` is set. The `integration` build tag can also be run directly with `go test -tags=integration ./...` after migrations.

## API docs and observability

Swagger UI is opt-in: set `ENABLE_SWAGGER=true` to serve `/swagger/index.html` and `/swagger/doc.json` outside production; Swagger is always disabled in production. The generated spec is committed under `docs/`.

Set `ENABLE_METRICS=true` for `GET /metrics`; restrict this endpoint to trusted monitoring systems in production. Security audits are structured JSON logs with request IDs and categorical context, not raw identifiers. Rate-limit denials/backend failures are also recorded in low-cardinality metrics. Grafana, Loki, and tracing infrastructure are not bundled.
