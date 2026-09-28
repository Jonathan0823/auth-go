# Authentication rate limiting

The HTTP adapter checks IP policies before calling auth services. Auth services check account/email/token policies so non-HTTP callers also get those limits. A request subject to both consumes the IP counter first, then the account/email/token counter. Each policy has a separate counter; requests denied by a counter are not incremented by that counter.

Counters use a **fixed window starting with the first request**. The next window starts on the first request after expiry. Keys are HMAC-SHA-256 digests of the policy, dimension, and identifier using `RATE_LIMIT_KEY`; raw emails, tokens, and IP addresses are not stored as rate-limit keys. Email keys are trimmed and lowercased.

## Default policies

Override a policy with `RATE_LIMIT_<NAME>=<limit>/<Go duration>` (for example, `RATE_LIMIT_LOGIN_IP=30/1m`). These defaults are defined in `internal/config/rate_limit.go`:

| Policy | Applies to | Default |
| --- | --- | --- |
| `login_ip` | Login HTTP requests, per client IP | 30/1m |
| `login_account` | Login attempts, per normalized email | 5/15m |
| `register_ip` | Registration HTTP requests, per client IP | 5/1h |
| `register_email` | Registration attempts, per normalized email | 5/1h |
| `recovery_ip` | Forgot/reset password HTTP requests, per client IP | 10/1h |
| `recovery_email` | Forgot-password attempts, per normalized email | 3/1h |
| `recovery_token` | Reset-password attempts, per token ID | 3/1h |
| `refresh_ip` | Refresh HTTP requests, per client IP | 30/1m |
| `verify_ip` | Resend-verification HTTP requests, per client IP | 10/1h |
| `verify_email` | Resend-verification attempts, per normalized email | 3/1h |
| `oauth_ip` | OAuth start and callback requests, per client IP | 10/1m |

A successful password login attempts to reset its `login_account` counter (reset failures do not fail the login). Other counters expire at the end of their windows. The email-verification link (`GET /api/auth/verify/email`) is **not** subject to the resend-verification policies.

## Select a backend

`RATE_LIMIT_BACKEND` defaults to `memory`. Starting the Compose Redis service does **not** change the backend.

```dotenv
RATE_LIMIT_BACKEND=redis
REDIS_ADDR=localhost:6379
REDIS_DB=0
RATE_LIMIT_KEY=replace-with-a-separate-random-secret-at-least-32-bytes
```

- `memory`: local to one process, capped by `RATE_LIMIT_MAX_MEMORY_KEYS` (default 10000). Counters disappear on restart, and multiple instances do not share them. Disabled in production unless `RATE_LIMIT_ALLOW_MEMORY_PRODUCTION=true` is explicitly set.
- `redis`: shared counters with atomic Lua updates and a TTL matching the policy window. Set `REDIS_ADDR`; `REDIS_PASSWORD` and `REDIS_DB` are optional. Use a dedicated Redis instance or logical database when isolation matters.
- `postgres`: shared counters in `rate_limit_buckets` (migration `000003`), serialized per key with a PostgreSQL advisory transaction lock. Expired rows are deleted during use and periodically.

Use Redis or PostgreSQL for multiple application instances. Set `RATE_LIMIT_KEY` separately from JWT, refresh-token, and session secrets; rotating it effectively starts fresh counters. Production validation requires it to be at least 32 bytes.

## IP trust and responses

Gin's `ClientIP()` trusts only the direct connection unless `TRUSTED_PROXIES` contains your actual proxy IPs or CIDRs (comma-separated). Do **not** trust arbitrary client-supplied forwarded headers; configure this only when traffic arrives through a known reverse proxy.

A denied request returns HTTP **429** with `Retry-After` (seconds) and `Cache-Control: no-store`. Backend errors fail closed with HTTP **503** (`authentication temporarily unavailable`), `Retry-After: 30`, and `Cache-Control: no-store`; there is no fallback to another backend. The retry header is clamped to 1–3600 seconds. The IP check lives in the HTTP adapter; the account/email/token error is mapped to the same responses by its error middleware. Denials and failures emit audit events/metrics without raw identifiers.
