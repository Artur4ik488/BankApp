# BankApp

BankApp is a small Go banking demo backed by PostgreSQL. It demonstrates account creation, deposits, withdrawals, and atomic transfers through a service and repository architecture.

Money is stored as an integer number of cents. For example, `25_00` in Go means 2,500 cents, or 25.00 in the chosen currency.

## Run the application

Requirements: Go, Docker, and Docker Compose.

```sh
test -f .env || cp .env.example .env
docker compose up -d
set -a
source .env
set +a
go run .
```

If `.env` already exists, make sure it also defines `DATABASE_URL` as shown in `.env.example`.

On a fresh PostgreSQL data directory, Docker runs the SQL file in `migrations/` automatically. The program then creates two accounts and performs this scenario:

1. Deposit 100.00 into Arthur's account.
2. Deposit 50.00 into Daniel's account.
3. Transfer 25.00 from Arthur to Daniel.
4. Withdraw 10.00 from Daniel.

The expected final balances are 75.00 for Arthur and 65.00 for Daniel. The program writes both balances as cents in its final JSON log entry.

## Test

```sh
go test ./...
go vet ./...
```
