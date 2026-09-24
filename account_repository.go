package main

import "context"

type AccountRepository interface {
	AddAccount(ctx context.Context, account Account) error
	GetAccount(ctx context.Context, id int) (Account, error)
	UpdateAccount(ctx context.Context, account Account) error
	UpdateAccounts(ctx context.Context, accountFrom, accountTo Account) error
	CreateAccount(ctx context.Context, account Account) (Account, error)
}
