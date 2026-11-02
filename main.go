package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := run(context.Background()); err != nil {
		slog.Error("application failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL is not set")
	}

	pool, err := NewPostgresPool(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	slog.Info("database connection established")

	repo := NewPostgresAccountRepository(pool)
	service := NewAccountService(repo)

	fromAccount, err := service.CreateAccount(ctx, "Arthur")
	if err != nil {
		return fmt.Errorf("create source account: %w", err)
	}

	toAccount, err := service.CreateAccount(ctx, "Daniel")
	if err != nil {
		return fmt.Errorf("create destination account: %w", err)
	}

	if err := service.Deposit(ctx, fromAccount.ID, 100_00); err != nil {
		return fmt.Errorf("deposit into source account: %w", err)
	}

	if err := service.Deposit(ctx, toAccount.ID, 50_00); err != nil {
		return fmt.Errorf("deposit into destination account: %w", err)
	}

	if err := service.Transfer(ctx, fromAccount.ID, toAccount.ID, 25_00); err != nil {
		return fmt.Errorf("transfer money: %w", err)
	}

	if err := service.Withdraw(ctx, toAccount.ID, 10_00); err != nil {
		return fmt.Errorf("withdraw from destination account: %w", err)
	}

	fromAccount, err = service.GetAccount(ctx, fromAccount.ID)
	if err != nil {
		return fmt.Errorf("get source account: %w", err)
	}

	toAccount, err = service.GetAccount(ctx, toAccount.ID)
	if err != nil {
		return fmt.Errorf("get destination account: %w", err)
	}

	slog.Info(
		"banking demo completed",
		"source_account", slog.GroupValue(
			slog.Int("id", fromAccount.ID),
			slog.String("owner", fromAccount.Owner),
			slog.Int64("balance_cents", int64(fromAccount.Balance)),
		),
		"destination_account", slog.GroupValue(
			slog.Int("id", toAccount.ID),
			slog.String("owner", toAccount.Owner),
			slog.Int64("balance_cents", int64(toAccount.Balance)),
		),
	)

	return nil
}
