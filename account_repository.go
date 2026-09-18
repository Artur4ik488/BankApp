package main

type AccountRepository interface {
	AddAccount(account Account) error
	GetAccount(id int) (Account, error)
	UpdateAccount(account Account) error
	UpdateAccounts(accountFrom, accountTo Account) error
}
