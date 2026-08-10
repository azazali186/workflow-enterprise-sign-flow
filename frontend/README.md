# SignFlow Console — Frontend

Premium admin dashboard for the SignFlow contract-signing backend.

## Stack

- **React 19 + Vite 8 + TypeScript**
- **Redux Toolkit** — auth session + UI state (sidebar, toasts)
- **TanStack React Query** — server state, caching, mutations
- **Tailwind CSS v4** — design tokens via `@theme` in `src/styles/index.css`
- **React Router 7** — lazy-loaded, permission-guarded routes

## Getting started

```bash
npm install
npm run dev        # http://localhost:5173 (proxies /api -> http://localhost:8080)
npm run build      # production build (code-split per page)
npm run preview
npm run typecheck  # tsc --noEmit
npm run lint       # oxlint
npm run test       # vitest (44 unit tests)
npm run test:coverage
```

Point `VITE_API_PROXY` at the backend when it isn't on `:8080`
(e.g. `VITE_API_PROXY=http://localhost:8080 npm run dev`). See `.env.example`
for the full list of variables.

## Backend API alignment

The app strictly follows the backend contract:

| Backend rule | Frontend handling |
|---|---|
| Envelope `{ code, message, data }` | Unwrapped by the API client (`src/services/api/client.ts`) |
| Errors `{ code, message }` | Normalized → toasts + inline field errors |
| POST/PATCH/DELETE only | All services use `post`/`patch`/`del` helpers |
| List endpoints: `POST /list` with `limit`, `cursor`, `filters`, `search`, `sort`, `date_from`, `date_to` | `useListController` + `useListQuery` |
| Cursor pagination (`next_cursor`, `has_more`, `total_count`) | `PaginationBar` — Previous/Next + "Showing 11–20 of 245" |
| Optional DB summary on lists | Rendered under page titles (e.g. status counts) |
| Bearer token | Attached by interceptor; 401 clears session (single sign-out) |

No mock data anywhere — every screen reads live endpoints.

## Structure

```
src/
  components/ui/        # Button, Input, Badge, Modal, Toast, DataTable, Skeleton…
  features/<entity>/    # pages + components per feature
  hooks/                # useListController, useListQuery, useToast, useSession, usePermission
  layouts/              # AppLayout, Sidebar, Header, navigation config
  services/             # api client + per-feature service files
  store/                # Redux slices (auth, ui) + toast middleware
  styles/index.css      # design tokens (Tailwind v4 @theme)
  types/                # api + per-domain entity types
  utils/                # cn, format, permissions
```

## Conventions

- **≤300 lines per file** (enforced; split types by domain to comply)
- No business logic inside UI components — hooks/services own behavior
- All lists: skeleton loaders, friendly empty states, error states with retry
- Role-based UI: sidebar items and routes gated by `permission`
  (`"METHOD /api/v1/path"`), `super_admin` bypasses
- Motion: 150–300ms, ease-out enters, `prefers-reduced-motion` respected
