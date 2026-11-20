package main

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
)

func TestDepositRejectsBalanceOverflow(t *testing.T) {
	account := Account{Balance: MaxMoney}

	err := account.Deposit(1)

	if !errors.Is(err, ErrBalanceLimit) {
		t.Fatalf("expected ErrBalanceLimit, got %v", err)
	}
	if account.Balance != MaxMoney {
		t.Fatalf("overflow changed balance to %d", account.Balance)
	}
}

func TestTransferRollsBackWhenDestinationHitsBalanceLimit(t *testing.T) {
	ctx := context.Background()
	store := NewAccountStore()
	service := NewAccountService(store)

	if err := store.AddAccount(ctx, Account{ID: 1, Owner: "Source", Balance: 100}); err != nil {
		t.Fatal(err)
	}
	if err := store.AddAccount(ctx, Account{ID: 2, Owner: "Destination", Balance: MaxMoney}); err != nil {
		t.Fatal(err)
	}

	err := service.Transfer(ctx, 1, 2, 50)
	if !errors.Is(err, ErrBalanceLimit) {
		t.Fatalf("expected ErrBalanceLimit, got %v", err)
	}

	source, err := store.GetAccount(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	destination, err := store.GetAccount(ctx, 2)
	if err != nil {
		t.Fatal(err)
	}
	if source.Balance != 100 || destination.Balance != MaxMoney {
		t.Fatalf("failed transfer changed balances: source=%d destination=%d", source.Balance, destination.Balance)
	}
}

func TestConcurrentDepositsAreNotLost(t *testing.T) {
	ctx := context.Background()
	store := NewAccountStore()
	service := NewAccountService(store)
	if err := store.AddAccount(ctx, Account{ID: 1, Owner: "Arthur"}); err != nil {
		t.Fatal(err)
	}

	const workers = 25
	const depositsPerWorker = 40
	var waitGroup sync.WaitGroup
	errorsChannel := make(chan error, workers)

	for range workers {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			for range depositsPerWorker {
				if err := service.Deposit(ctx, 1, 1); err != nil {
					errorsChannel <- err
					return
				}
			}
		}()
	}
	waitGroup.Wait()
	close(errorsChannel)
	for err := range errorsChannel {
		t.Fatalf("concurrent deposit failed: %v", err)
	}

	account, err := store.GetAccount(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	wantBalance := Money(workers * depositsPerWorker)
	if account.Balance != wantBalance {
		t.Fatalf("expected balance %d, got %d", wantBalance, account.Balance)
	}
}

func TestCreateAccountValidatesOwner(t *testing.T) {
	service := NewAccountService(NewAccountStore())

	tests := []string{
		strings.Repeat("a", MaxOwnerLength+1),
		"Arthur\nAdmin",
		"\t",
	}
	for _, owner := range tests {
		if _, err := service.CreateAccount(context.Background(), owner); !errors.Is(err, ErrInvalidOwner) {
			t.Fatalf("expected owner %q to be rejected, got %v", owner, err)
		}
	}
}

func TestStoreHonorsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	store := NewAccountStore()
	if _, err := store.ListAccounts(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}
