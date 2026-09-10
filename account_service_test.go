package main

import "testing"

func TestSuccessfulDeposit(t *testing.T) {

	store := NewAccountStore()
	service := NewAccountService(store)

	testAccount := Account{
		ID:      1,
		Owner:   "Arthur",
		Balance: 50.00,
	}

	if err := store.AddAccount(testAccount); err != nil {
		t.Fatal(err)
	}

	if err := service.Deposit(1, 100); err != nil {
		t.Fatal(err)
	}

	account, err := store.GetAccount(1)
	if err != nil {
		t.Fatal(err)
	}

	if account.Balance != 150 {
		t.Fatalf("excepted balance 150 got %v", account.Balance)
	}

}

func TestAccountDoesNotExist(t *testing.T) {

	store := NewAccountStore()
	service := NewAccountService(store)

	if err := service.Deposit(1, 100); err == nil {
		t.Fatal("excepted error for non-existing account")
	}

}

func TestInvalidDeposit(t *testing.T) {

	store := NewAccountStore()
	service := NewAccountService(store)

	testAccount := Account{
		ID:      1,
		Owner:   "Arthur",
		Balance: 50.00,
	}

	if err := store.AddAccount(testAccount); err != nil {
		t.Fatal(err)
	}

	if err := service.Deposit(1, -100); err == nil {
		t.Fatal("excepted error for uncorrect value of deposit")
	}

	account, err := store.GetAccount(1)
	if err != nil {
		t.Fatal(err)
	}

	if account.Balance != 50 {
		t.Fatalf("excepted initial balance 50 got %v", account.Balance)
	}

}
