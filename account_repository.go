package main

import "context"

type AccountRepository interface {
	Health(ctx context.Context) error
	ListAccounts(ctx context.Context) ([]Account, error)
	GetAccount(ctx context.Context, id int) (Account, error)
	CreateAccount(ctx context.Context, account Account) (Account, error)
	Deposit(ctx context.Context, accountID int, amount Money) error
	Withdraw(ctx context.Context, accountID int, amount Money) error
	Transfer(ctx context.Context, fromAccountID, toAccountID int, amount Money) error
}
