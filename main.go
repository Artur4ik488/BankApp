package main

import (
	"errors"
	"fmt"
)

func main() {

	account1 := Account{
		ID:      1,
		Owner:   "Arthur",
		Balance: 300,
	}

	err := account1.Withdraw(400)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(account1.Balance)

}

type Account struct {
	ID      int
	Owner   string
	Balance float64
}

func findAccountByID(accounts []Account, id int) (*Account, bool) {
	for i, account := range accounts {
		if account.ID == id {
			return &accounts[i], true
		}
	}
	return nil, false
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
