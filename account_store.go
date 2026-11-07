package main

import (
	"context"
	"sort"
	"sync"
)

type AccountStore struct {
	mu       sync.RWMutex
	accounts map[int]Account
	nextID   int
}

var _ AccountRepository = (*AccountStore)(nil)

func NewAccountStore() *AccountStore {
	accountStore := make(map[int]Account)
	return &AccountStore{
		accounts: accountStore,
		nextID:   1,
	}
}

// CREAETE ACCOUNT //

func (store *AccountStore) CreateAccount(
	ctx context.Context,
	account Account,
) (Account, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	account.ID = store.nextID
	if err := store.addAccount(account); err != nil {
		return Account{}, err
	}
	return account, nil
}

// CREAETE ACCOUNT //

// ADD ACCOUNT //

func (store *AccountStore) AddAccount(
	ctx context.Context,
	account Account,
) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	return store.addAccount(account)
}

func (store *AccountStore) addAccount(account Account) error {
	_, ok := store.accounts[account.ID]
	if ok {
		return ErrAccountExists
	}
	store.accounts[account.ID] = account

	if account.ID >= store.nextID {
		store.nextID = account.ID + 1
	}
	return nil
}

// ADD ACCOUNT //

// GET ACCOUNT //

func (store *AccountStore) GetAccount(
	ctx context.Context,
	id int,
) (Account, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	value, ok := store.accounts[id]

	if !ok {
		return Account{}, ErrAccountNotFound
	}
	return value, nil
}

func (store *AccountStore) ListAccounts(ctx context.Context) ([]Account, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	accounts := make([]Account, 0, len(store.accounts))
	for _, account := range store.accounts {
		accounts = append(accounts, account)
	}
	sort.Slice(accounts, func(i, j int) bool {
		return accounts[i].ID < accounts[j].ID
	})
	return accounts, nil
}

// GET ACCOUNT //

// UPDATE //

func (store *AccountStore) UpdateAccount(
	ctx context.Context,
	account Account,
) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	_, ok := store.accounts[account.ID]

	if !ok {
		return ErrAccountNotFound
	}

	store.accounts[account.ID] = account
	return nil
}

func (store *AccountStore) UpdateAccounts(
	ctx context.Context,
	accountFrom, accountTo Account,
) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	_, okFrom := store.accounts[accountFrom.ID]
	if !okFrom {
		return ErrAccountNotFound
	}
	_, okTo := store.accounts[accountTo.ID]
	if !okTo {
		return ErrAccountNotFound
	}

	store.accounts[accountFrom.ID] = accountFrom
	store.accounts[accountTo.ID] = accountTo
	return nil

}

// UPDATE //
