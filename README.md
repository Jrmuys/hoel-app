# hoel-app

Household Operations Dashboard monorepo.

## Workspace Layout

- `frontend/` SvelteKit UI (planned)
- `backend/` Go API, schedulers, and SQLite cache
- `infra/` Optional containerization assets
- `docs/` Design and implementation notes

## Prerequisites

- Go `1.24+`
- Node `22.12+`
- npm `10+`

## MVP Scope (Phase 1)

See `docs/mvp-scope.md` for committed MVP boundaries and deferred features.

## Run Locally (Native-First)

Backend:

```sh
cd backend
cp .env.example .env
go test ./...
go run ./cmd/server
```

Frontend:

```sh
cd frontend
cp .env.example .env
npm install
npm run check
npm run dev
```

## Run with Docker Compose (Optional)

```sh
docker compose -f infra/docker-compose.yml up --build
```

## Development Mode

Native-first local development is the default. Docker support is optional for parity and deployment.
