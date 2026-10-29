package main

import (
	"context"
	"log/slog"
	"os"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		slog.Error("database URL is empty")
		os.Exit(1)
	}

	ctx := context.Background()

	pool, err := NewPostgresPool(ctx, databaseURL)
	if err != nil {
		slog.Error("failed to create database pool", "error", err)
		os.Exit(1)
	}

	defer pool.Close()

	slog.Info("database connection established", "address", "localhost:5433")

	repo := NewPostgresAccountRepository(pool)

	account := Account{
		Owner:   "GigaMegaArthur",
		Balance: 0,
	}

	createdAccount, err := repo.CreateAccount(ctx, account)
	if err != nil {
		slog.Error("failed to create account", "error", err)
		os.Exit(0)
	}

	slog.Info(
		"account created",
		"id", createdAccount.ID,
		"owner", createdAccount.Owner,
		"balance", createdAccount.Balance,
	)
}
