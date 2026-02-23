# Environment Variables

## Backend

| Variable | Required | Default | Purpose |
| --- | --- | --- | --- |
| `APP_HOST` | No | `127.0.0.1` | API bind host |
| `APP_PORT` | No | `8080` | API bind port |
| `APP_READ_TIMEOUT` | No | `10s` | HTTP server read timeout |
| `APP_WRITE_TIMEOUT` | No | `15s` | HTTP server write timeout |
| `APP_SHUTDOWN_TIMEOUT` | No | `10s` | Graceful shutdown budget |
| `APP_ALLOWED_ORIGINS` | No | `http://localhost:5173,http://127.0.0.1:5173` | Comma-separated CORS origins |
| `SQLITE_PATH` | No | `./hoel.db` | SQLite database file |
| `MIGRATIONS_DIR` | No | `./migrations` | SQL migration directory |
| `OUTBOUND_HTTP_TIMEOUT` | No | `8s` | HTTP timeout for outbound integrations |
| `OUTBOUND_RETRY_COUNT` | No | `2` | Retries for retryable outbound failures |
| `OUTBOUND_RETRY_BACKOFF` | No | `300ms` | Backoff between outbound retries |
| `PGHST_ENDPOINT` | Yes* | _(empty)_ | PGH.st endpoint to poll (`/api/daily-operations` remains empty until set) |
| `PGHST_POLL_INTERVAL` | No | `12h` | Polling interval for PGH.st sync |

\* Required to enable PGH polling.

## Frontend

| Variable | Required | Default | Purpose |
| --- | --- | --- | --- |
| `PUBLIC_API_BASE_URL` | No | `http://127.0.0.1:8080` | API base URL consumed by browser code |

## Upcoming Integrations (Next Slices)

- TickTick OAuth: `TICKTICK_CLIENT_ID`, `TICKTICK_CLIENT_SECRET`, `TICKTICK_REDIRECT_URI`
- Telegram alerts: `TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID`
