package main

import (
	"errors"
)

type AccountStore struct {
	accounts map[int]Account
}

func NewAccountStore() *AccountStore {
	accountStore := make(map[int]Account)
	return &AccountStore{
		accounts: accountStore,
	}
}

func (store *AccountStore) AddAccount(account Account) error {

	_, ok := store.accounts[account.ID]

	if ok {
		return errors.New("Account with current ID is existing already.")
	}
	store.accounts[account.ID] = account
	return nil
}

func (store *AccountStore) GetAccount(id int) (Account, error) {

	value, ok := store.accounts[id]

	if !ok {
		return Account{}, errors.New("Account does not exist by this ID")
	}
	return value, nil

}
