package main

import "errors"

type Account struct {
	ID      int
	Owner   string
	Balance float64
}

func (account *Account) Deposit(amount float64) error {
	if amount <= 0 {
		return errors.New("deposit amount must be greater than zero")
	}
	account.Balance += amount
	return nil
}

func (account *Account) Withdraw(amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	if amount > account.Balance {
		return errors.New("insufficient funds")
	}
	account.Balance -= amount
	return nil
}
