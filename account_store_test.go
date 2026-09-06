package main

import (
	"testing"
)

func TestAddAndGetAccount(t *testing.T) {

	testAccount := Account{
		ID:      1,
		Owner:   "Arthur",
		Balance: 50.99,
	}

	store := NewAccountStore()

	if err := store.AddAccount(testAccount); err != nil {
		t.Fatalf("failed to add account %v", err)
	}

	account, err2 := store.GetAccount(1)
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
		Balance: 50.99,
	}

	store := NewAccountStore()

	testAccount2 := Account{
		ID:      1,
		Owner:   "Daniel",
		Balance: 60.99,
	}

	err := store.AddAccount(testAccount)
	if err != nil {
		t.Fatalf("The first account wasn't added!")
	}

	err2 := store.AddAccount(testAccount2)
	if err2 == nil {
		t.Fatalf("Duplicate was allowed!")
	}

}

func TestGetNonExistingAccount(t *testing.T) {

	store := NewAccountStore()

	_, err := store.GetAccount(999)
	if err == nil {
		t.Fatalf("Method Get does not work correctly")
	}

}
