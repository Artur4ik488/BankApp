package main

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
		return ErrAccountExists
	}
	store.accounts[account.ID] = account
	return nil
}

func (store *AccountStore) GetAccount(id int) (Account, error) {

	value, ok := store.accounts[id]

	if !ok {
		return Account{}, ErrAccountNotFound
	}
	return value, nil

}

// UPDATE
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

// UPDATE
