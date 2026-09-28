# Documentation

- [Setup and operations](setup.md) — configuration, running, migrations, tests, observability, and API documentation.
- [Authentication and API](authentication.md) — tokens, cookies, account flows, and routes.
- [Rate limiting](rate-limiting.md) — policies, backends, proxy trust, and failure handling.
- [Storage and caching](storage-and-caching.md) — PostgreSQL vs Redis and the current absence of a data cache.
- [Architecture](architecture.md) — package boundaries and where to add new functionality.

The generated [Swagger specification](swagger.yaml) describes the HTTP contract; `make swagger` regenerates it from Go annotations.
