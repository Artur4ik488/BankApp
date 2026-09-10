package main

import (
	"errors"
)

type AccountService struct {
	store *AccountStore
}

func NewAccountService(store *AccountStore) *AccountService {
	return &AccountService{
		store: store,
	}
}

func (store *AccountStore) UpdateAccount(account Account) error {

	_, ok := store.accounts[account.ID]

	if !ok {
		return errors.New("account doesn't exist")
	}

	store.accounts[account.ID] = account
	return nil
}

func (service *AccountService) Deposit(accountID int, amount float64) error {

	account, err1 := service.store.GetAccount(accountID)
	if err1 != nil {
		return err1
	}

	if err := account.Deposit(amount); err != nil {
		return err
	}

	if err := service.store.UpdateAccount(account); err != nil {
		return err
	}
	return nil

}
