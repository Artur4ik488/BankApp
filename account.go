package main

type Money int64

type Account struct {
	ID      int
	Owner   string
	Balance Money
}

func (account *Account) Deposit(amount Money) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	account.Balance += amount
	return nil
}

func (account *Account) Withdraw(amount Money) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if amount > account.Balance {
		return ErrInsufficientFunds
	}
	account.Balance -= amount
	return nil
}
