package main

type AccountStore struct {
	accounts map[int]Account
	nextID   int
}

func NewAccountStore() *AccountStore {
	accountStore := make(map[int]Account)
	return &AccountStore{
		accounts: accountStore,
		nextID:   1,
	}
}

// CREAETE ACCOUNT //

func (store *AccountStore) CreateAccount(account Account) (Account, error) {

	account.ID = store.nextID

	if err := store.AddAccount(account); err != nil {
		return Account{}, err
	}
	return account, nil
}

// CREAETE ACCOUNT //

// ADD ACCOUNT //

func (store *AccountStore) AddAccount(account Account) error {

	_, ok := store.accounts[account.ID]

	if ok {
		return ErrAccountExists
	}
	store.accounts[account.ID] = account

	if account.ID >= store.nextID {
		store.nextID = store.nextID + 1
	}
	return nil
}

// ADD ACCOUNT //

// GET ACCOUNT //

func (store *AccountStore) GetAccount(id int) (Account, error) {

	value, ok := store.accounts[id]

	if !ok {
		return Account{}, ErrAccountNotFound
	}
	return value, nil

}

// GET ACCOUNT //

// UPDATE //

func (store *AccountStore) UpdateAccount(account Account) error {

	_, ok := store.accounts[account.ID]

	if !ok {
		return ErrAccountNotFound
	}

	store.accounts[account.ID] = account
	return nil
}

func (store *AccountStore) UpdateAccounts(accountFrom, accountTo Account) error {

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
