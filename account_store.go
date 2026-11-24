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

func (store *AccountStore) Health(ctx context.Context) error {
	return ctx.Err()
}

func (store *AccountStore) CreateAccount(
	ctx context.Context,
	account Account,
) (Account, error) {
	if err := ctx.Err(); err != nil {
		return Account{}, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()

	account.ID = store.nextID
	if err := store.addAccount(account); err != nil {
		return Account{}, err
	}
	return account, nil
}

func (store *AccountStore) AddAccount(
	ctx context.Context,
	account Account,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
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
func (store *AccountStore) GetAccount(
	ctx context.Context,
	id int,
) (Account, error) {
	if err := ctx.Err(); err != nil {
		return Account{}, err
	}
	store.mu.RLock()
	defer store.mu.RUnlock()

	value, ok := store.accounts[id]

	if !ok {
		return Account{}, ErrAccountNotFound
	}
	return value, nil
}

func (store *AccountStore) ListAccounts(ctx context.Context) ([]Account, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
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

func (store *AccountStore) Deposit(
	ctx context.Context,
	accountID int,
	amount Money,
) (Account, error) {
	if err := ctx.Err(); err != nil {
		return Account{}, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()

	account, ok := store.accounts[accountID]
	if !ok {
		return Account{}, ErrAccountNotFound
	}
	if err := account.Deposit(amount); err != nil {
		return Account{}, err
	}
	store.accounts[accountID] = account
	return account, nil
}

func (store *AccountStore) Withdraw(
	ctx context.Context,
	accountID int,
	amount Money,
) (Account, error) {
	if err := ctx.Err(); err != nil {
		return Account{}, err
	}
	store.mu.Lock()
	defer store.mu.Unlock()

	account, ok := store.accounts[accountID]
	if !ok {
		return Account{}, ErrAccountNotFound
	}
	if err := account.Withdraw(amount); err != nil {
		return Account{}, err
	}
	store.accounts[accountID] = account
	return account, nil
}

func (store *AccountStore) Transfer(
	ctx context.Context,
	fromAccountID, toAccountID int,
	amount Money,
) (Account, Account, error) {
	if err := ctx.Err(); err != nil {
		return Account{}, Account{}, err
	}
	if fromAccountID == toAccountID {
		return Account{}, Account{}, ErrSameAccount
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	accountFrom, ok := store.accounts[fromAccountID]
	if !ok {
		return Account{}, Account{}, ErrAccountNotFound
	}
	accountTo, ok := store.accounts[toAccountID]
	if !ok {
		return Account{}, Account{}, ErrAccountNotFound
	}
	if err := accountFrom.Withdraw(amount); err != nil {
		return Account{}, Account{}, err
	}
	if err := accountTo.Deposit(amount); err != nil {
		return Account{}, Account{}, err
	}

	store.accounts[fromAccountID] = accountFrom
	store.accounts[toAccountID] = accountTo
	return accountFrom, accountTo, nil
}
