package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/001_create_accounts.sql
var accountsMigration string

var integrationSchemaSequence atomic.Uint64

func TestPostgresConcurrentDepositsAreNotLost(t *testing.T) {
	repo := newPostgresIntegrationRepository(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	account := createIntegrationAccount(t, repo, "Deposits", 0)

	const operations = 64
	errorsChannel := make(chan error, operations)
	var waitGroup sync.WaitGroup
	for range operations {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			updated, err := repo.Deposit(ctx, account.ID, 1)
			if err == nil && updated.ID != account.ID {
				err = fmt.Errorf("deposit returned account %d, want %d", updated.ID, account.ID)
			}
			errorsChannel <- err
		}()
	}
	waitGroup.Wait()
	close(errorsChannel)
	assertNoOperationErrors(t, errorsChannel)

	stored, err := repo.GetAccount(ctx, account.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Balance != operations {
		t.Fatalf("expected balance %d, got %d", operations, stored.Balance)
	}
}

func TestPostgresConcurrentWithdrawalsNeverProduceNegativeBalance(t *testing.T) {
	repo := newPostgresIntegrationRepository(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	account := createIntegrationAccount(t, repo, "Withdrawals", 40)

	const operations = 80
	var successful atomic.Int64
	var rejected atomic.Int64
	errorsChannel := make(chan error, operations)
	var waitGroup sync.WaitGroup
	for range operations {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			updated, err := repo.Withdraw(ctx, account.ID, 1)
			switch {
			case err == nil:
				if updated.Balance < 0 {
					errorsChannel <- fmt.Errorf("withdrawal returned negative balance %d", updated.Balance)
					return
				}
				successful.Add(1)
			case errors.Is(err, ErrInsufficientFunds):
				rejected.Add(1)
			default:
				errorsChannel <- err
				return
			}
			errorsChannel <- nil
		}()
	}
	waitGroup.Wait()
	close(errorsChannel)
	assertNoOperationErrors(t, errorsChannel)

	if successful.Load() != 40 || rejected.Load() != 40 {
		t.Fatalf("expected 40 successful and 40 rejected withdrawals, got %d and %d", successful.Load(), rejected.Load())
	}
	stored, err := repo.GetAccount(ctx, account.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Balance != 0 {
		t.Fatalf("expected zero balance, got %d", stored.Balance)
	}
}

func TestPostgresOppositeDirectionTransfersDoNotDeadlock(t *testing.T) {
	repo := newPostgresIntegrationRepository(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	accountA := createIntegrationAccount(t, repo, "Account A", 1_000)
	accountB := createIntegrationAccount(t, repo, "Account B", 1_000)

	const operationsPerDirection = 40
	errorsChannel := make(chan error, operationsPerDirection*2)
	var waitGroup sync.WaitGroup
	for range operationsPerDirection {
		waitGroup.Add(2)
		go runIntegrationTransfer(ctx, &waitGroup, errorsChannel, repo, accountA.ID, accountB.ID, 1)
		go runIntegrationTransfer(ctx, &waitGroup, errorsChannel, repo, accountB.ID, accountA.ID, 1)
	}

	done := make(chan struct{})
	go func() {
		waitGroup.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
		t.Fatalf("opposite-direction transfers did not finish: %v", ctx.Err())
	}
	close(errorsChannel)
	assertNoOperationErrors(t, errorsChannel)

	assertIntegrationBalances(t, repo, map[int]Money{
		accountA.ID: 1_000,
		accountB.ID: 1_000,
	})
}

func TestPostgresFailedTransferChangesNeitherAccount(t *testing.T) {
	repo := newPostgresIntegrationRepository(t)
	ctx := context.Background()
	accountFrom := createIntegrationAccount(t, repo, "Source", 50)
	accountTo := createIntegrationAccount(t, repo, "Destination", 20)

	_, _, err := repo.Transfer(ctx, accountFrom.ID, accountTo.ID, 100)
	if !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("expected ErrInsufficientFunds, got %v", err)
	}
	assertIntegrationBalances(t, repo, map[int]Money{
		accountFrom.ID: 50,
		accountTo.ID:   20,
	})
}

func TestPostgresConcurrentTransfersPreserveTotalMoney(t *testing.T) {
	repo := newPostgresIntegrationRepository(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	accounts := make([]Account, 4)
	for i := range accounts {
		accounts[i] = createIntegrationAccount(t, repo, fmt.Sprintf("Account %d", i+1), 10_000)
	}

	const operations = 240
	errorsChannel := make(chan error, operations)
	var waitGroup sync.WaitGroup
	for i := range operations {
		from := accounts[i%len(accounts)]
		to := accounts[(i%len(accounts)+i%3+1)%len(accounts)]
		waitGroup.Add(1)
		go runIntegrationTransfer(ctx, &waitGroup, errorsChannel, repo, from.ID, to.ID, 1)
	}
	waitGroup.Wait()
	close(errorsChannel)
	assertNoOperationErrors(t, errorsChannel)

	var total Money
	for _, account := range accounts {
		stored, err := repo.GetAccount(ctx, account.ID)
		if err != nil {
			t.Fatal(err)
		}
		total += stored.Balance
	}
	if total != 40_000 {
		t.Fatalf("expected total balance 40000, got %d", total)
	}
}

func TestPostgresMissingAccountsReturnNotFound(t *testing.T) {
	repo := newPostgresIntegrationRepository(t)
	ctx := context.Background()
	existing := createIntegrationAccount(t, repo, "Existing", 100)

	if _, err := repo.Deposit(ctx, 999_999, 1); !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("deposit: expected ErrAccountNotFound, got %v", err)
	}
	if _, err := repo.Withdraw(ctx, 999_999, 1); !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("withdraw: expected ErrAccountNotFound, got %v", err)
	}
	if _, _, err := repo.Transfer(ctx, 999_999, existing.ID, 1); !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("missing source: expected ErrAccountNotFound, got %v", err)
	}
	if _, _, err := repo.Transfer(ctx, existing.ID, 999_999, 1); !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("missing destination: expected ErrAccountNotFound, got %v", err)
	}
	assertIntegrationBalances(t, repo, map[int]Money{existing.ID: 100})
}

func TestPostgresMaxMoneyFailureRollsBackTransaction(t *testing.T) {
	repo := newPostgresIntegrationRepository(t)
	ctx := context.Background()
	accountFrom := createIntegrationAccount(t, repo, "Source", 100)
	accountTo := createIntegrationAccount(t, repo, "At limit", MaxMoney)

	if _, err := repo.Deposit(ctx, accountTo.ID, 1); !errors.Is(err, ErrBalanceLimit) {
		t.Fatalf("deposit: expected ErrBalanceLimit, got %v", err)
	}
	if _, _, err := repo.Transfer(ctx, accountFrom.ID, accountTo.ID, 1); !errors.Is(err, ErrBalanceLimit) {
		t.Fatalf("transfer: expected ErrBalanceLimit, got %v", err)
	}
	assertIntegrationBalances(t, repo, map[int]Money{
		accountFrom.ID: 100,
		accountTo.ID:   MaxMoney,
	})
}

func TestPostgresMoneyOperationsReturnCommittedAccounts(t *testing.T) {
	repo := newPostgresIntegrationRepository(t)
	ctx := context.Background()
	accountFrom := createIntegrationAccount(t, repo, "Source", 100)
	accountTo := createIntegrationAccount(t, repo, "Destination", 25)

	deposited, err := repo.Deposit(ctx, accountFrom.ID, 50)
	if err != nil {
		t.Fatal(err)
	}
	if deposited.Balance != 150 {
		t.Fatalf("expected deposited balance 150, got %d", deposited.Balance)
	}

	withdrawn, err := repo.Withdraw(ctx, accountFrom.ID, 20)
	if err != nil {
		t.Fatal(err)
	}
	if withdrawn.Balance != 130 {
		t.Fatalf("expected withdrawn balance 130, got %d", withdrawn.Balance)
	}

	updatedFrom, updatedTo, err := repo.Transfer(ctx, accountFrom.ID, accountTo.ID, 30)
	if err != nil {
		t.Fatal(err)
	}
	if updatedFrom.Balance != 100 || updatedTo.Balance != 55 {
		t.Fatalf("unexpected returned balances: from=%d to=%d", updatedFrom.Balance, updatedTo.Balance)
	}
	assertIntegrationBalances(t, repo, map[int]Money{
		accountFrom.ID: 100,
		accountTo.ID:   55,
	})
}

func newPostgresIntegrationRepository(t *testing.T) *PostgresAccountRepository {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping PostgreSQL integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	adminPool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create admin pool: %v", err)
	}
	t.Cleanup(adminPool.Close)
	if err := adminPool.Ping(ctx); err != nil {
		t.Fatalf("ping PostgreSQL: %v", err)
	}

	schemaName := fmt.Sprintf(
		"bankapp_test_%d_%d",
		time.Now().UnixNano(),
		integrationSchemaSequence.Add(1),
	)
	quotedSchema := pgx.Identifier{schemaName}.Sanitize()
	if _, err := adminPool.Exec(ctx, "CREATE SCHEMA "+quotedSchema); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		if _, err := adminPool.Exec(cleanupCtx, "DROP SCHEMA "+quotedSchema+" CASCADE"); err != nil {
			t.Errorf("drop test schema: %v", err)
		}
	})

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse test database URL: %v", err)
	}
	if config.ConnConfig.RuntimeParams == nil {
		config.ConnConfig.RuntimeParams = make(map[string]string)
	}
	config.ConnConfig.RuntimeParams["search_path"] = schemaName
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("create schema pool: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping schema pool: %v", err)
	}
	if _, err := pool.Exec(ctx, accountsMigration); err != nil {
		t.Fatalf("apply accounts migration: %v", err)
	}

	return NewPostgresAccountRepository(pool)
}

func createIntegrationAccount(
	t *testing.T,
	repo *PostgresAccountRepository,
	owner string,
	balance Money,
) Account {
	t.Helper()
	account, err := repo.CreateAccount(context.Background(), Account{
		Owner:   owner,
		Balance: balance,
	})
	if err != nil {
		t.Fatalf("create account: %v", err)
	}
	return account
}

func runIntegrationTransfer(
	ctx context.Context,
	waitGroup *sync.WaitGroup,
	errorsChannel chan<- error,
	repo *PostgresAccountRepository,
	fromID, toID int,
	amount Money,
) {
	defer waitGroup.Done()
	_, _, err := repo.Transfer(ctx, fromID, toID, amount)
	errorsChannel <- err
}

func assertNoOperationErrors(t *testing.T, errorsChannel <-chan error) {
	t.Helper()
	for err := range errorsChannel {
		if err != nil {
			t.Fatalf("concurrent operation failed: %v", err)
		}
	}
}

func assertIntegrationBalances(
	t *testing.T,
	repo *PostgresAccountRepository,
	expected map[int]Money,
) {
	t.Helper()
	for accountID, expectedBalance := range expected {
		account, err := repo.GetAccount(context.Background(), accountID)
		if err != nil {
			t.Fatalf("get account %d: %v", accountID, err)
		}
		if account.Balance != expectedBalance {
			t.Fatalf("account %d: expected balance %d, got %d", accountID, expectedBalance, account.Balance)
		}
	}
}
