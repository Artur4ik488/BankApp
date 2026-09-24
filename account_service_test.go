package main

import (
	"context"
	"errors"
	"testing"
)

// DEPOSIT TESTS

func TestSuccessfulDeposit(t *testing.T) {
	store := NewAccountStore()
	service := NewAccountService(store)

	ctx := context.Background()

	testAccount := Account{
		ID:      1,
		Owner:   "Arthur",
		Balance: 50_00,
	}

	if err := store.AddAccount(ctx, testAccount); err != nil {
		t.Fatalf("failed to add account: %v", err)
	}

	if err := service.Deposit(ctx, 1, 100_00); err != nil {
		t.Fatalf("deposit failed: %v", err)
	}

	account, err := store.GetAccount(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get account: %v", err)
	}

	if account.Balance != 150_00 {
		t.Fatalf("expected balance 15000, got %v", account.Balance)
	}
}

func TestDepositAccountNotFound(t *testing.T) {
	store := NewAccountStore()
	service := NewAccountService(store)

	ctx := context.Background()

	err := service.Deposit(ctx, 999, 100_00)

	if !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
}

func TestInvalidDeposit(t *testing.T) {
	store := NewAccountStore()
	service := NewAccountService(store)

	ctx := context.Background()

	testAccount := Account{
		ID:      1,
		Owner:   "Arthur",
		Balance: 50_00,
	}

	if err := store.AddAccount(ctx, testAccount); err != nil {
		t.Fatalf("failed to add account: %v", err)
	}

	err := service.Deposit(ctx, 1, -100_00)

	if !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("expected ErrInvalidAmount, got %v", err)
	}

	account, err := store.GetAccount(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get account: %v", err)
	}

	if account.Balance != 50_00 {
		t.Fatalf("expected balance 5000, got %v", account.Balance)
	}
}

// WITHDRAW TESTS

func TestSuccessfulWithdraw(t *testing.T) {
	store := NewAccountStore()
	service := NewAccountService(store)

	testAccount := Account{
		ID:      1,
		Owner:   "Arthur",
		Balance: 150_00,
	}

	ctx := context.Background()

	if err := store.AddAccount(ctx, testAccount); err != nil {
		t.Fatalf("failed to add account: %v", err)
	}

	if err := service.Withdraw(ctx, 1, 100_00); err != nil {
		t.Fatalf("withdraw failed: %v", err)
	}

	account, err := store.GetAccount(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get account: %v", err)
	}

	if account.Balance != 50_00 {
		t.Fatalf("expected balance 5000, got %v", account.Balance)
	}
}

func TestWithdrawInsufficientFunds(t *testing.T) {
	store := NewAccountStore()
	service := NewAccountService(store)

	ctx := context.Background()

	testAccount := Account{
		ID:      1,
		Owner:   "Arthur",
		Balance: 150_00,
	}

	if err := store.AddAccount(ctx, testAccount); err != nil {
		t.Fatalf("failed to add account: %v", err)
	}

	err := service.Withdraw(ctx, 1, 200_00)

	if !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("expected ErrInsufficientFunds, got %v", err)
	}

	account, err := store.GetAccount(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get account: %v", err)
	}

	if account.Balance != 150_00 {
		t.Fatalf("expected balance 15000, got %v", account.Balance)
	}
}

func TestInvalidWithdraw(t *testing.T) {
	store := NewAccountStore()
	service := NewAccountService(store)

	ctx := context.Background()

	testAccount := Account{
		ID:      1,
		Owner:   "Arthur",
		Balance: 150_00,
	}

	if err := store.AddAccount(ctx, testAccount); err != nil {
		t.Fatalf("failed to add account: %v", err)
	}

	err := service.Withdraw(ctx, 1, -100_00)

	if !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("expected ErrInvalidAmount, got %v", err)
	}

	account, err := store.GetAccount(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get account: %v", err)
	}

	if account.Balance != 150_00 {
		t.Fatalf("expected balance 15000, got %v", account.Balance)
	}
}

func TestWithdrawAccountNotFound(t *testing.T) {
	store := NewAccountStore()
	service := NewAccountService(store)

	ctx := context.Background()

	err := service.Withdraw(ctx, 999, 100_00)

	if !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
}

// TRANSFER TESTS

func TestSuccessfulTransfer(t *testing.T) {
	store := NewAccountStore()
	service := NewAccountService(store)

	ctx := context.Background()

	accountFrom := Account{
		ID:      1,
		Owner:   "Arthur",
		Balance: 150_00,
	}

	accountTo := Account{
		ID:      2,
		Owner:   "Daniel",
		Balance: 200_00,
	}

	if err := store.AddAccount(ctx, accountFrom); err != nil {
		t.Fatalf("failed to add source account: %v", err)
	}

	if err := store.AddAccount(ctx, accountTo); err != nil {
		t.Fatalf("failed to add destination account: %v", err)
	}

	if err := service.Transfer(ctx, 1, 2, 5000); err != nil {
		t.Fatalf("transfer failed: %v", err)
	}

	updatedFrom, err := store.GetAccount(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get source account: %v", err)
	}

	updatedTo, err := store.GetAccount(ctx, 2)
	if err != nil {
		t.Fatalf("failed to get destination account: %v", err)
	}

	if updatedFrom.Balance != 100_00 {
		t.Fatalf("expected source balance 10000, got %v", updatedFrom.Balance)
	}

	if updatedTo.Balance != 250_00 {
		t.Fatalf("expected destination balance 25000, got %v", updatedTo.Balance)
	}
}

func TestTransferSameAccount(t *testing.T) {
	store := NewAccountStore()
	service := NewAccountService(store)

	ctx := context.Background()

	err := service.Transfer(ctx, 1, 1, 100_00)

	if !errors.Is(err, ErrSameAccount) {
		t.Fatalf("expected ErrSameAccount, got %v", err)
	}
}

func TestTransferSourceAccountNotFound(t *testing.T) {
	store := NewAccountStore()
	service := NewAccountService(store)

	ctx := context.Background()

	accountTo := Account{
		ID:      2,
		Owner:   "Daniel",
		Balance: 200_00,
	}

	if err := store.AddAccount(ctx, accountTo); err != nil {
		t.Fatalf("failed to add destination account: %v", err)
	}

	err := service.Transfer(ctx, 1, 2, 50_00)

	if !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
}

func TestTransferDestinationAccountNotFound(t *testing.T) {
	store := NewAccountStore()
	service := NewAccountService(store)

	ctx := context.Background()

	accountFrom := Account{
		ID:      1,
		Owner:   "Arthur",
		Balance: 150_00,
	}

	if err := store.AddAccount(ctx, accountFrom); err != nil {
		t.Fatalf("failed to add source account: %v", err)
	}

	err := service.Transfer(ctx, 1, 2, 50_00)

	if !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
}

func TestTransferInsufficientFunds(t *testing.T) {
	store := NewAccountStore()
	service := NewAccountService(store)

	ctx := context.Background()

	accountFrom := Account{
		ID:      1,
		Owner:   "Arthur",
		Balance: 50_00,
	}

	accountTo := Account{
		ID:      2,
		Owner:   "Daniel",
		Balance: 200_00,
	}

	if err := store.AddAccount(ctx, accountFrom); err != nil {
		t.Fatalf("failed to add source account: %v", err)
	}

	if err := store.AddAccount(ctx, accountTo); err != nil {
		t.Fatalf("failed to add destination account: %v", err)
	}

	err := service.Transfer(ctx, 1, 2, 100_00)

	if !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("expected ErrInsufficientFunds, got %v", err)
	}

	updatedFrom, err := store.GetAccount(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get source account: %v", err)
	}

	updatedTo, err := store.GetAccount(ctx, 2)
	if err != nil {
		t.Fatalf("failed to get destination account: %v", err)
	}

	if updatedFrom.Balance != 50_00 {
		t.Fatalf("expected source balance 5000, got %v", updatedFrom.Balance)
	}

	if updatedTo.Balance != 200_00 {
		t.Fatalf("expected destination balance 20000, got %v", updatedTo.Balance)
	}
}

func TestTransferInvalidAmount(t *testing.T) {
	store := NewAccountStore()
	service := NewAccountService(store)

	ctx := context.Background()

	accountFrom := Account{
		ID:      1,
		Owner:   "Arthur",
		Balance: 150_00,
	}

	accountTo := Account{
		ID:      2,
		Owner:   "Daniel",
		Balance: 200_00,
	}

	if err := store.AddAccount(ctx, accountFrom); err != nil {
		t.Fatalf("failed to add source account: %v", err)
	}

	if err := store.AddAccount(ctx, accountTo); err != nil {
		t.Fatalf("failed to add destination account: %v", err)
	}

	err := service.Transfer(ctx, 1, 2, -50_00)

	if !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("expected ErrInvalidAmount, got %v", err)
	}
}

func TestCreateAccountSuccessful(t *testing.T) {

	store := NewAccountStore()
	service := NewAccountService(store)

	ctx := context.Background()

	account, err := service.CreateAccount(ctx, "Arthur")
	if err != nil {
		t.Fatal(err)
	}

	storedAccount, err := store.GetAccount(ctx, account.ID)
	if err != nil {
		t.Fatal(err)
	}
	if storedAccount != account {
		t.Fatal("stored account does not match created account")
	}

	if account.ID != 1 {
		t.Fatalf("expected ID 1 got %v", account.ID)
	}
	if account.Balance != 0 {
		t.Fatalf("expected Balance 0 got %v", account.Balance)
	}
	if account.Owner != "Arthur" {
		t.Fatalf("expected owner \"Arthur\" got %v", account.Owner)
	}

}

func TestCreateAccountThatStartsWithID7ThenNextMustToBe7(t *testing.T) {
	store := NewAccountStore()
	service := NewAccountService(store)

	ctx := context.Background()

	account := Account{
		ID:      6,
		Owner:   "Daniel",
		Balance: 0,
	}

	if err := service.store.AddAccount(ctx, account); err != nil {
		t.Fatal(err)
	}

	account2, err := service.CreateAccount(ctx, "Arthur")
	if err != nil {
		t.Fatal(err)
	}

	if account2.ID != 7 {
		t.Fatalf("ID of account doesn't match, must be 7 got %v", account2.ID)
	}
	storedAccount, err := store.GetAccount(ctx, account2.ID)
	if err != nil {
		t.Fatal(err)
	}
	if storedAccount != account2 {
		t.Fatal("stored account does not match created account2")
	}
}

func TestCreateTwoAccountsSuccessful(t *testing.T) {
	store := NewAccountStore()
	service := NewAccountService(store)

	ctx := context.Background()

	account1, err1 := service.CreateAccount(ctx, "Arthur")
	if err1 != nil {
		t.Fatal(err1)
	}

	storedAccount1, err2 := store.GetAccount(ctx, account1.ID)
	if err2 != nil {
		t.Fatal(err2)
	}
	if storedAccount1 != account1 {
		t.Fatal("stored account does not match created account1")
	}

	account2, err3 := service.CreateAccount(ctx, "Daniel")
	if err3 != nil {
		t.Fatal(err3)
	}

	storedAccount2, err4 := store.GetAccount(ctx, account2.ID)
	if err4 != nil {
		t.Fatal(err4)
	}
	if storedAccount2 != account2 {
		t.Fatal("stored account does not match created accoun2")
	}

	if account1.ID != 1 {
		t.Fatalf("expected ID 1 got %v", account1.ID)
	}
	if account1.Balance != 0 {
		t.Fatalf("expected Balance 0 got %v", account1.Balance)
	}
	if account1.Owner != "Arthur" {
		t.Fatalf("expected owner \"Arthur\" got %v", account1.Owner)
	}

	if account2.ID != 2 {
		t.Fatalf("expected ID 2 got %v", account2.ID)
	}
	if account2.Balance != 0 {
		t.Fatalf("expected Balance 0 got %v", account2.Balance)
	}
	if account2.Owner != "Daniel" {
		t.Fatalf("expected owner \"Daniel\" got %v", account2.Owner)
	}
}

func TestCreateAccountInvalidOwner(t *testing.T) {

	store := NewAccountStore()
	service := NewAccountService(store)

	ctx := context.Background()

	_, err := service.CreateAccount(ctx, "")

	if !errors.Is(err, ErrInvalidOwner) {
		t.Fatalf("expected ErrInvalidOwner got %v", err)
	}

}
