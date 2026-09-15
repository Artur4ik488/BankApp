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

// TRANSFER TESTS //

func TestSuccessfulAccountTransfer(t *testing.T) {

	store := NewAccountStore()
	service := NewAccountService(store)

	testAccount1 := Account{
		ID:      1,
		Owner:   "Arthur",
		Balance: 150.00,
	}

	testAccount2 := Account{
		ID:      2,
		Owner:   "Daniel",
		Balance: 200.00,
	}

	if err := store.AddAccount(testAccount1); err != nil {
		t.Fatal(err)
	}
	if err := store.AddAccount(testAccount2); err != nil {
		t.Fatal(err)
	}

	if err := service.Transfer(1, 2, 50); err != nil {
		t.Fatal(err)
	}

	account1, err := store.GetAccount(1)
	if err != nil {
		t.Fatal(err)
	}

	account2, err := store.GetAccount(2)
	if err != nil {
		t.Fatal(err)
	}

	if account1.Balance != 100 || account2.Balance != 250 {
		t.Fatalf("expected account1 balance 100 got %v\nexpected account2 balance 250 got %v", account1.Balance, account2.Balance)
	}

}

func TestTheSameAccounts(t *testing.T) {

	store := NewAccountStore()
	service := NewAccountService(store)

	if err := service.Transfer(1, 1, 100); err == nil {
		t.Fatal("expected error for the same accounts")
	}
}

// TRANSFER TESTS //
