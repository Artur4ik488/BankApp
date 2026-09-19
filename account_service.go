package main

// STRUCT

type AccountService struct {
	store AccountRepository
}

// SRTUCT

// NEW SERVICE

func NewAccountService(store AccountRepository) *AccountService {
	return &AccountService{
		store: store,
	}
}

// NEW SERVICE

// NEW ACCOUNT //
func (service *AccountService) CreateAccount(owner string) (Account, error) {

	if owner == "" {
		return Account{}, ErrInvalidOwner
	}

	account := Account{
		Owner:   owner,
		Balance: 0,
	}

	return service.store.CreateAccount(account)

}

// NEW ACCOUNT //

// DEPOSIT
func (service *AccountService) Deposit(accountID int, amount Money) error {

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

// DEPOSIT

// WITHDRAW
func (service *AccountService) Withdraw(accountID int, amount Money) error {

	account, err1 := service.store.GetAccount(accountID)
	if err1 != nil {
		return err1
	}

	if err := account.Withdraw(amount); err != nil {
		return err
	}

	if err := service.store.UpdateAccount(account); err != nil {
		return err
	}
	return nil

}

// WITHDRAW

// TRANSFER //

func (service *AccountService) Transfer(fromAccountID, toAccountID int, amount Money) error {

	if fromAccountID == toAccountID {
		return ErrSameAccount
	}

	accountFrom, err := service.store.GetAccount(fromAccountID)
	if err != nil {
		return err
	}
	accountTo, err := service.store.GetAccount(toAccountID)
	if err != nil {
		return err
	}

	if err := accountFrom.Withdraw(amount); err != nil {
		return err
	}

	if err := accountTo.Deposit(amount); err != nil {
		return err
	}

	if err := service.store.UpdateAccounts(accountFrom, accountTo); err != nil {
		return err
	}
	return nil

}

// TRANSFER //
