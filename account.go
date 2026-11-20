package main

type Money int64

// MaxMoney stays below JavaScript's largest exactly representable integer so
// balances can travel through the JSON API without losing cents.
const MaxMoney Money = 9_000_000_000_000_000

type Account struct {
	ID      int    `json:"id"`
	Owner   string `json:"owner"`
	Balance Money  `json:"balance"`
}

func (account *Account) Deposit(amount Money) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if account.Balance < 0 || amount > MaxMoney || account.Balance > MaxMoney-amount {
		return ErrBalanceLimit
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
