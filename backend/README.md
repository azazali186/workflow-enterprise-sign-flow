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
| `POST /api/v1/health`  | Liveness check (public)          |

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
- `docs/swagger.json` / `/swagger` — full OpenAPI spec of all 48 endpoints
