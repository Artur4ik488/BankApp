package main

import "context"

type AccountService struct {
	store AccountRepository
}

func NewAccountService(store AccountRepository) *AccountService {
	return &AccountService{
		store: store,
	}
}

func (service *AccountService) CreateAccount(
	ctx context.Context,
	owner string,
) (Account, error) {

	if owner == "" {
		return Account{}, ErrInvalidOwner
	}

	account := Account{
		Owner:   owner,
		Balance: 0,
	}

	return service.store.CreateAccount(ctx, account)
}

func (service *AccountService) GetAccount(
	ctx context.Context,
	accountID int,
) (Account, error) {
	return service.store.GetAccount(ctx, accountID)
}

func (service *AccountService) Deposit(
	ctx context.Context,
	accountID int,
	amount Money,
) error {

	account, err := service.store.GetAccount(ctx, accountID)
	if err != nil {
		return err
	}

	if err := account.Deposit(amount); err != nil {
		return err
	}

	if err := service.store.UpdateAccount(ctx, account); err != nil {
		return err
	}
	return nil
}

func (service *AccountService) Withdraw(
	ctx context.Context,
	accountID int,
	amount Money,
) error {

	account, err := service.store.GetAccount(ctx, accountID)
	if err != nil {
		return err
	}

	if err := account.Withdraw(amount); err != nil {
		return err
	}

	if err := service.store.UpdateAccount(ctx, account); err != nil {
		return err
	}
	return nil
}

func (service *AccountService) Transfer(
	ctx context.Context,
	fromAccountID, toAccountID int,
	amount Money,
) error {

	if fromAccountID == toAccountID {
		return ErrSameAccount
	}

	accountFrom, err := service.store.GetAccount(ctx, fromAccountID)
	if err != nil {
		return err
	}
	accountTo, err := service.store.GetAccount(ctx, toAccountID)
	if err != nil {
		return err
	}

	if err := accountFrom.Withdraw(amount); err != nil {
		return err
	}

	if err := accountTo.Deposit(amount); err != nil {
		return err
	}

	if err := service.store.UpdateAccounts(ctx, accountFrom, accountTo); err != nil {
		return err
	}
	return nil
}
