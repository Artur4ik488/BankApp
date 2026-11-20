# BankApp

BankApp is a small full-stack banking dashboard built with Go, PostgreSQL, and plain HTML/CSS/JavaScript. It supports account creation, deposits, withdrawals, and atomic transfers through a service and repository architecture.

Money is stored as an integer number of cents. For example, `25_00` in Go means 2,500 cents, or 25.00 in the chosen currency.

## Run the application

Requirements: Go, Docker, and Docker Compose.

```sh
test -f .env || cp .env.example .env
docker compose up -d --wait
set -a
source .env
set +a
go run .
```

BankApp builds its connection URL from `POSTGRES_DB`, `POSTGRES_USER`, and `POSTGRES_PASSWORD`. `POSTGRES_HOST` and `POSTGRES_PORT` default to `localhost` and `5433`. A standalone `DATABASE_URL` is also supported when those PostgreSQL variables are not set.

The application listens on `127.0.0.1:8080` by default. Set `HOST` or `PORT` to override either value.

On a fresh PostgreSQL data directory, Docker runs the SQL file in `migrations/` automatically. Open [http://localhost:8080](http://localhost:8080) after the application starts.

For a quick temporary demo without Docker or PostgreSQL:

```sh
BANKAPP_STORE=memory go run .
```

The in-memory mode resets all accounts when the application stops.

## Features

- Responsive account dashboard for desktop and mobile
- Create and list bank accounts
- Deposit and withdraw money
- Transfer money atomically between accounts
- JSON API with validation and structured errors
- Graceful HTTP server shutdown
- Embedded frontend: the Go binary serves all web assets

## API

| Method | Route | Purpose |
| --- | --- | --- |
| `GET` | `/api/health` | Check service availability |
| `GET` | `/api/accounts` | List accounts |
| `POST` | `/api/accounts` | Create an account |
| `GET` | `/api/accounts/{id}` | Get an account |
| `POST` | `/api/accounts/{id}/deposit` | Deposit money |
| `POST` | `/api/accounts/{id}/withdraw` | Withdraw money |
| `POST` | `/api/transfers` | Transfer between accounts |

API amounts are integer cents. For example, `2500` represents `$25.00`.

## Troubleshooting

- **Port already in use:** stop the previous BankApp process with `Ctrl+C`, or run `PORT=8081 go run .`.
- **Database authentication failed:** update `POSTGRES_PASSWORD` in `.env`. Existing PostgreSQL data directories keep the password used when they were first created.
- **Database is unavailable:** run `docker compose up -d --wait` and confirm that `docker compose ps` reports PostgreSQL as healthy.
- **Temporary demo:** run `BANKAPP_STORE=memory go run .` to use BankApp without PostgreSQL. Data is cleared when the process stops.

## Test

```sh
go test ./...
go test -race ./...
go vet ./...
```
