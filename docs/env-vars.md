# Environment Variables

## Backend

| Variable | Required | Default | Purpose |
| --- | --- | --- | --- |
| `APP_HOST` | No | `127.0.0.1` | API bind host |
| `APP_PORT` | No | `8080` | API bind port |
| `APP_READ_TIMEOUT` | No | `10s` | HTTP server read timeout |
| `APP_WRITE_TIMEOUT` | No | `15s` | HTTP server write timeout |
| `APP_SHUTDOWN_TIMEOUT` | No | `10s` | Graceful shutdown budget |
| `SQLITE_PATH` | No | `./hoel.db` | SQLite database file |
| `MIGRATIONS_DIR` | No | `./migrations` | SQL migration directory |

## Frontend

| Variable | Required | Default | Purpose |
| --- | --- | --- | --- |
| `PUBLIC_API_BASE_URL` | No | `http://127.0.0.1:8080` | API base URL consumed by browser code |

## Upcoming Integrations (Next Slices)

- TickTick OAuth: `TICKTICK_CLIENT_ID`, `TICKTICK_CLIENT_SECRET`, `TICKTICK_REDIRECT_URI`
- Telegram alerts: `TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID`
