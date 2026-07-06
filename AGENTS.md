# AURA Development Guide

## Project Overview

AURA (Automated Utility for Retrieval of Assets) is a web app for managing MediUX image assets for Plex/Emby/Jellyfin. It has a **Go backend** and **Next.js frontend**, deployed via Docker.

## Quick Commands

### Frontend (from `frontend/`)
```bash
npm run dev          # Dev server (proxies API to :8888)
npm run build        # Production build
npm run lint         # ESLint (must pass with 0 warnings)
npm run lint:fix     # Auto-fix lint issues
npm run typecheck    # TypeScript check (tsc --noEmit)
npm run format       # Prettier format (auto-runs in pre-commit)
npm run format:check # Check formatting without writing
```

### Backend (from `backend/`)
```bash
go build -o main .   # Build binary
go run .             # Run locally (listens on :8888)
```

### API Docs
```bash
cd backend && sh generate_go_docs.sh  # Regenerate Swagger docs (requires swag CLI)
```

### Docker
```bash
docker compose up    # Start both services
# Frontend: :3000  |  Backend API: :8888
```

## Architecture

### Backend (`backend/`)
- **Entry**: `main.go` → `startup.go` (bootstrap → preflight → warmup pipeline)
- **Router**: `routing/router.go` uses `go-chi/chi/v5` with JWT auth middleware
- **Routes**: `routing/routes.go` — two modes:
  - **Onboarding mode** (config invalid): limited routes for initial setup
  - **Full mode** (config valid): all `/api/*` routes behind JWT auth
- **Config**: YAML-based, loaded from `/config` dir at runtime. Structs in `config/config.go`
- **Database**: SQLite via `mattn/go-sqlite3`. Migrations in `database/migration/`
- **State**: Atomic router swap pattern — `activeHandler` stores the current router, swapped when app fully loads
- **Port**: Backend serves on `:8888` by default

### Frontend (`frontend/`)
- **Framework**: Next.js 16 + React 19 + Tailwind CSS 4
- **UI**: shadcn/ui (new-york style) with Radix primitives, Lucide icons
- **State**: Zustand stores persisted to IndexedDB via `localforage` (`src/lib/stores/`)
- **API**: Axios client (`src/services/api-client.ts`) with auto 401 redirect to `/login`
- **Path alias**: `@/*` → `./src/*`
- **Proxy**: Next.js rewrites `/api/*` → `localhost:8888/api/*` (see `next.config.ts`)
- **Theme**: Class-based dark mode (default: dark), `ThemeProvider` wraps app
- **Layout**: Gabarito font, sonner toasts, Navbar + Footer

### Key Integration Points
- Backend JWT token stored in browser `localStorage` as `aura-auth-token`
- Frontend reads `NEXT_PUBLIC_APP_VERSION` from `VERSION.txt` at build time
- Swagger UI served at `/swagger/*` (auto-generated from Go annotations)
- WebSocket connections for Plex event listening and auto-download

## Code Style

### Frontend
- Prettier: 2-space indent, double quotes, trailing commas (es5), 120 char width, LF line endings
- Import order enforced by `@trivago/prettier-plugin-sort-imports`: react → next → @/components → @/lib → @/hooks → @/types → relative
- ESLint: type-aware rules for `src/**/*.{ts,tsx}` using `tsconfig.eslint.json`
- Unused vars prefixed with `_` are allowed

### Backend
- Go with `go-chi` router, zerolog for logging
- Error handling uses custom error type with `Message` field (check `Err.Message != ""` for errors)
- Config uses YAML tags for serialization

## Git Hooks (Conventional Commits)

**Commit format**: `type(scope): subject`
- Types: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert
- Examples: `feat(frontend): add user sorting`, `fix: handle nil cache item`
- Hooks run from `.githooks/` — ensure `git config core.hooksPath .githooks`

**Pre-commit hook** runs:
1. Frontend lint + typecheck + format (only when `frontend/` files change)
2. Swagger doc regeneration (only when backend routing/config/models change)

## Gotchas

- **No test suite**: There are currently no Go tests or frontend tests in the repo
- **Two-phase startup**: Backend starts with onboarding routes first, swaps to full routes after config validation + DB init. If config is invalid, only onboarding endpoints are available
- **Standalone output**: Frontend builds with `output: "standalone"` for Docker deployment
- **Port mismatch**: Frontend dev server is `:3000`, backend API is `:8888` — the proxy in `next.config.ts` bridges them
- **React strict mode**: Disabled (`reactStrictMode: false` in next.config.ts)
