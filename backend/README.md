# AeroXe SignFlow Backend

Production-grade REST backend for a digital signature and contract
management platform. Built with **Hertz**, **GORM** (no raw SQL),
**Redis**, and **NATS**.

## Conventions

- **HTTP methods:** only `POST`, `PATCH`, `DELETE` — no `PUT`, no `GET`.
- **Request/response:** always `snake_case` JSON in the body. No path
  variables and no query parameters; record ids are sent in the body.
- **Pagination:** every list/report endpoint uses server-side cursor
  pagination with `{limit, cursor, filters, search, sort, date_from,
  date_to, date_field}` and returns a pagination summary + DB summary.
- **IDs:** UUID v7 primary keys.
- **RBAC:** permissions are seeded automatically from registered routes at
  startup; public routes (login, health, swagger, metrics, ws) are skipped.
- **Audit:** before/after change tracking; no sensitive data is logged.
- **Security headers + CORS:** hardened headers on every response;
  `CORS_ALLOWED_ORIGINS` allow-list (empty = same-origin, `*` = dev only).
- **Request IDs:** every request (public routes included) carries an
  `X-Request-ID` echoed back and correlated through logs and audit entries.
- **Readiness:** `POST /api/v1/health` pings Postgres and Redis and reports
  `503` when a dependency is down.
- **Resilient NATS:** the outbox relay never silently drops events — while
  NATS is unreachable it keeps reconnecting in the background and events
  stay pending until delivered.

## Quick Start

```bash
# 1. Start infrastructure (Postgres, Redis, NATS)
docker compose up -d postgres redis nats

# 2. Configure environment
cp .env.example .env

# 3. Apply migrations
make migrate

# 4. Run the API (default :8080)
make run
```

## Useful Endpoints

| Endpoint               | Purpose                          |
|------------------------|----------------------------------|
| `POST /api/v1/auth/login` | Obtain a bearer token         |
| `POST /api/v1/auth/me`    | Current user + roles          |
| `/swagger`             | Interactive Swagger UI           |
| `/metrics`             | Prometheus metrics               |
| `POST /api/v1/health`  | Liveness + readiness check (public) |

## Commands

```bash
make build       # compile server + migrate binaries
make test        # run tests
make test-race   # tests with the race detector
make vet         # static analysis
make swagger     # regenerate OpenAPI docs
make infra-up    # docker compose up
```

## Auth & RBAC

1. `POST /api/v1/auth/login` with `{email, password}` returns
   `access_token` (JWT) plus the user.
2. Send `Authorization: Bearer <token>` on every API call.
3. The bootstrap admin (see `ADMIN_EMAIL`/`ADMIN_PASSWORD`) is seeded with
   the system `super_admin` role, which holds every permission.
4. Roles and permissions are managed via `/api/v1/roles/*` and
   `/api/v1/permissions/*`.

## Testing

Unit and integration tests cover the pagination cursor logic, encryption
serializer, cache, outbox relay and the auth flow (SQLite in-memory):

```bash
make test
```

## Documentation

- `docs/ARCHITECTURE.md` — high-level architecture and design decisions
- `docs/swagger.json` / `/swagger` — full OpenAPI spec of all 60 API endpoints
  (plus 5 public routes: login, health, swagger UI, metrics, WebSocket)

## Development on Windows

The project targets 64-bit builds (Hertz/sonic does not compile on 32-bit
toolchains). The `Makefile` already forces `GOARCH=amd64`, so prefer
`make run` / `make test`. If you invoke `go` directly, use
`GOARCH=amd64 go run ./cmd/server` — the default 32-bit Go install will fail
with a `sonic … overflows int` compile error.
