# Architecture

```text
cmd/             application entrypoint
internal/
  bootstrap/      dependency construction and resource lifecycle
  config/         environment loading and validation
  core/
    domain/       business models and errors
    port/         application-owned contracts and inputs
    ratelimit/    transport-independent rate-limit coordination
    service/      authentication and user use cases
  adapter/
    inbound/http/ Gin handlers, routing, DTOs, and HTTP middleware
    outbound/     PostgreSQL, rate-limit stores, JWT, OAuth, password, and email
  observability/  logging, metrics, and audit implementations
docs/             generated Swagger files and hand-written guides
migrations/       PostgreSQL schema migrations
```

Dependencies point inward: core code does not import HTTP, configuration, observability, or infrastructure adapters. `bootstrap` is the composition root that wires application-owned ports to concrete adapters and closes resources. The application decides *when* to apply account/email/token policies; HTTP owns client IP extraction and HTTP responses; stores implement persistence. See [Rate limiting](rate-limiting.md).

JWT encoding and token hashing live in the outbound JWT adapter, not in the core service. HTTP request/response DTOs are mapped explicitly rather than exposing domain objects directly. If a future workflow needs a queue, its publisher belongs in an outbound adapter and its worker in an inbound adapter; this template does not implement a queue or outbox.

Go source and test filenames under `cmd/` and `internal/` are lowercase and descriptive, with underscores between words. Test files append `_test.go`; tagged integration tests use `_integration_test.go`. Generator-controlled filenames remain unchanged. The test in `internal/architecture/` checks that core dependencies stay inward.
