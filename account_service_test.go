package main

import "testing"

// DEPOSIT TESTS
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
		t.Fatalf("expected balance 150 got %v", account.Balance)
	}

}

func TestAccountDoesNotExistForDeposit(t *testing.T) {

	store := NewAccountStore()
	service := NewAccountService(store)

	if err := service.Deposit(1, 100); err == nil {
		t.Fatal("expected error for non-existing account")
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
		t.Fatal("expected error for uncorrect value of deposit")
	}

	account, err := store.GetAccount(1)
	if err != nil {
		t.Fatal(err)
	}

	if account.Balance != 50 {
		t.Fatalf("expected initial balance 50 got %v", account.Balance)
	}

}

// DEPOSIT TESTS //

// WITHDRAW TESTS //
func TestSuccessfulWithdraw(t *testing.T) {

	store := NewAccountStore()
	service := NewAccountService(store)

	testAccount := Account{
		ID:      1,
		Owner:   "Arthur",
		Balance: 150.00,
	}

	if err := store.AddAccount(testAccount); err != nil {
		t.Fatal(err)
	}

	if err := service.Withdraw(1, 100); err != nil {
		t.Fatal(err)
	}

	account, err := store.GetAccount(1)
	if err != nil {
		t.Fatal(err)
	}

	if account.Balance != 50 {
		t.Fatalf("expected balance 50 got %v", account.Balance)
	}

}

func TestNotEnoughMoney(t *testing.T) {

	store := NewAccountStore()
	service := NewAccountService(store)

	testAccount := Account{
		ID:      1,
		Owner:   "Arthur",
		Balance: 150.00,
	}

	if err := store.AddAccount(testAccount); err != nil {
		t.Fatal(err)
	}

	if err := service.Withdraw(1, 200); err == nil {
		t.Fatal("expected error for not enogh money")
	}

	account, err := store.GetAccount(1)
	if err != nil {
		t.Fatal(err)
	}

	if account.Balance != 150 {
		t.Fatalf("expected balance 150 got %v", account.Balance)
	}

}

func TestInvalidWithdraw(t *testing.T) {
	store := NewAccountStore()
	service := NewAccountService(store)

	testAccount := Account{
		ID:      1,
		Owner:   "Arthur",
		Balance: 150.00,
	}

	if err := store.AddAccount(testAccount); err != nil {
		t.Fatal(err)
	}

	if err := service.Withdraw(1, -100); err == nil {
		t.Fatal("expected error for invalid amount")
	}
}

func TestAccountDoesNotExistForWithdraw(t *testing.T) {
	store := NewAccountStore()
	service := NewAccountService(store)

	if err := service.Withdraw(1, 100); err == nil {
		t.Fatal(`expected error for non-existing account`)
	}

}

// WITHDRAW TESTS //
