# SignFlow Backend — High-Level Architecture

## Overview

SignFlow is a digital signature and contract management platform. The backend
is a single Go service built on **Hertz** (HTTP), **GORM** (data), **Redis**
(cache/session/rate-limit), and **NATS** (event streaming). It exposes a
REST API under `/api/v1` with strict conventions:

- **Methods:** only `POST`, `PATCH`, `DELETE` (no `PUT`, no `GET`).
- **Payloads:** all input/output in `snake_case` JSON bodies. No path
  variables, no query parameters — every request carries its data (including
  record ids) in the body.
- **Pagination:** server-side cursor pagination on every list/report endpoint,
  with filters, dynamic-column sorting, date-range filtering and a
  pagination + DB summary in every response.

## Component Map

```
                    ┌──────────────────────────────────────────────┐
                    │                Hertz HTTP server              │
                    │  Recovery → RequestLog → RateLimit → Auth →   │
                    │  RBAC → handlers                              │
                    └──────────┬───────────────────────┬───────────┘
                               │                       │
                    ┌──────────▼──────────┐   ┌────────▼───────────┐
                    │  Services (modules)  │   │ WebSocket hub (/ws)│
                    │  auth, contracts, …  │   └────────┬───────────┘
                    └──────────┬──────────┘            │ events
                               │                       │
                    ┌──────────▼──────────┐   ┌────────▼───────────┐
                    │ Generic GORM repo   │   │ Event bus → NATS   │
                    │  (cache-aware)      │   │  via outbox relay  │
                    └──────────┬──────────┘   └────────┬───────────┘
                               │                       │
                ┌──────────────▼──────────┐   ┌────────▼───────────┐
                │  PostgreSQL (GORM)      │   │ Redis              │
                │  AutoMigrate + seed     │   │ cache/locks/session│
                └─────────────────────────┘   └────────────────────┘
```

## Key Design Decisions

### 1. RBAC with self-seeded permissions
At startup the **route registry** walks every registered handler, builds
`(method, path)` permission rows and upserts them into the `permissions`
table (mirroring `storeNewPermissions` from the reference implementation).
Routes marked `PUBLIC` (login, health, swagger, metrics, ws) are excluded
from auth/RBAC. Users get roles, roles get permissions, and the RBAC
middleware checks `(method, path)` of each request against the union of the
user's role permissions. Guards are cached in Redis and invalidated on role
changes.

### 2. Outbox pattern for reliable events
Any domain transition (contract created, signature captured, verification
verified) enqueues an outbox row **in the same DB transaction** as the
business change. The outbox relay polls pending rows, publishes them to NATS,
marks them published, and retries failures with backoff before marking the
event `failed`. This gives at-least-once delivery without dual-write problems.

### 3. Cursor pagination that is DB-agnostic
Lists accept `{limit, cursor, filters, search, sort, date_from, date_to,
date_field}`. The repository:

- computes `total_count` on the filtered query (no cursor),
- applies the cursor predicate `(sort_col < v OR (sort_col = v AND id < i))`
  or the `>`/`>` variant for ascending order,
- orders by `(sort_col, id)` and fetches `limit + 1` rows to derive
  `has_more` and `next_cursor`,
- runs an optional summary query (status breakdowns, totals) on the same
  filtered query.

Time columns are compared via `strftime('%s', …)` (SQLite) or
`EXTRACT(EPOCH FROM …)` (Postgres) so cursors are stable across dialects.

### 4. Security posture
- Passwords: bcrypt. OTP codes: SHA-256 hashed, never logged.
- Sensitive columns (`object_key`, certificate `data`) are transparently
  encrypted at rest with AES-256-GCM via a GORM serializer.
- JWT single-session enforcement: a server-side session hash in Redis is
  compared on every request (single sign-out).
- Audit logs record `before`/`after` changes and never contain secrets;
  logins are logged with success/failure.
- Rate limiting (token bucket in Redis), request body size limit, recovery
  middleware, and UUID v7 primary keys throughout.

### 5. Observability
- Prometheus metrics (`/metrics`): request duration, rate-limit and RBAC
  rejections, outbox backlog/failures, circuit breaker state.
- Structured zap logs with request ids.
- Swagger UI at `/swagger` with every endpoint annotated.

## Reliability Patterns

| Pattern            | Where                                                     |
|--------------------|-----------------------------------------------------------|
| Circuit breaker    | `internal/pkg/breaker` — guards NATS/cache round trips    |
| Retry w/ backoff   | `internal/pkg/retryutil` — outbox publishes, DB connect   |
| Outbox             | `internal/outbox` — transactional outbox + relay          |
| Distributed lock   | `internal/lock` — redsync over Redis (idempotent jobs)    |
| Body size limits   | Hertz `MaxRequestBodySize` + per-route guard              |
| Metrics            | `internal/pkg/metrics` — Prometheus                       |
| Distributed cache  | `internal/cache` — Redis (memory fallback in dev)         |
| Graceful shutdown  | SIGINT/SIGTERM → drain + close                            |

## Directory Layout

```
cmd/server       entrypoint: wiring, middleware, route registration, seed
cmd/migrate      standalone AutoMigrate runner
internal/config  env-based configuration
internal/database GORM dialector, serializer registration, model list
internal/cache   Redis + in-memory cache (unified interface)
internal/natsx   NATS connection + subject helpers
internal/lock    distributed lock
internal/outbox  transactional outbox + relay worker
internal/events  in-process event bus + NATS publisher
internal/audit   audit + login log writers
internal/middleware auth, RBAC, rate limit, request log, recovery
internal/registry route registry + permission seeding
internal/seed    bootstrap admin + system roles/permissions
internal/ws      WebSocket hub
internal/modules/*  one folder per entity (model + service + handler)
internal/pkg/*   shared libraries (repo, pagination, jwt, crypto, …)
docs             generated Swagger 2.0 (docs.go / swagger.json / yaml)
```

## Local Development

```bash
docker compose up -d postgres redis nats   # infra
cp .env.example .env                       # adjust secrets
make migrate                               # apply schema
make run                                   # start API on :8080
make swagger                               # regenerate OpenAPI docs
make test                                  # run the suite
```

Open http://localhost:8080/swagger for the interactive API docs and
http://localhost:8080/metrics for Prometheus metrics.
