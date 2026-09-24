package main

import (
	"context"
	"errors"
	"testing"
)

func TestAddAndGetAccount(t *testing.T) {

	testAccount := Account{
		ID:      1,
		Owner:   "Arthur",
		Balance: 50_99,
	}

	store := NewAccountStore()
	ctx := context.Background()

	if err := store.AddAccount(ctx, testAccount); err != nil {
		t.Fatalf("failed to add account %v", err)
	}

	account, err2 := store.GetAccount(ctx, 1)
	if err2 != nil {
		t.Fatalf("failed to get account %v", err2)
	}

	if account != testAccount {
		t.Errorf("Accounts are misaligned")
	}

}

func TestAddDuplicateAccount(t *testing.T) {

	testAccount := Account{
		ID:      1,
		Owner:   "Arthur",
		Balance: 50_99,
	}

	store := NewAccountStore()
	ctx := context.Background()

	testAccount2 := Account{
		ID:      1,
		Owner:   "Daniel",
		Balance: 60_99,
	}

	err := store.AddAccount(ctx, testAccount)
	if err != nil {
		t.Fatalf("The first account wasn't added!")
	}

	if err := store.AddAccount(ctx, testAccount2); !errors.Is(err, ErrAccountExists) {
		t.Fatalf("Duplicate was allowed!")
	}

}

func TestGetNonExistingAccount(t *testing.T) {

	store := NewAccountStore()

	ctx := context.Background()

	if _, err := store.GetAccount(ctx, 999); !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("Method Get does not work correctly")
	}

}
