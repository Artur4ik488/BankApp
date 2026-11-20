package main

import (
	"context"
	"strings"
	"unicode"
	"unicode/utf8"
)

type AccountService struct {
	store AccountRepository
}

const MaxOwnerLength = 80

func NewAccountService(store AccountRepository) *AccountService {
	return &AccountService{
		store: store,
	}
}

func (service *AccountService) CreateAccount(
	ctx context.Context,
	owner string,
) (Account, error) {
	owner = strings.TrimSpace(owner)
	if owner == "" || utf8.RuneCountInString(owner) > MaxOwnerLength || strings.IndexFunc(owner, unicode.IsControl) >= 0 {
		return Account{}, ErrInvalidOwner
	}

	account := Account{
		Owner:   owner,
		Balance: 0,
	}

	return service.store.CreateAccount(ctx, account)
}

func (service *AccountService) Health(ctx context.Context) error {
	return service.store.Health(ctx)
}

func (service *AccountService) ListAccounts(ctx context.Context) ([]Account, error) {
	return service.store.ListAccounts(ctx)
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
	if amount <= 0 {
		return ErrInvalidAmount
	}
	return service.store.Deposit(ctx, accountID, amount)
}

func (service *AccountService) Withdraw(
	ctx context.Context,
	accountID int,
	amount Money,
) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	return service.store.Withdraw(ctx, accountID, amount)
}

func (service *AccountService) Transfer(
	ctx context.Context,
	fromAccountID, toAccountID int,
	amount Money,
) error {
	if fromAccountID == toAccountID {
		return ErrSameAccount
	}
	if amount <= 0 {
		return ErrInvalidAmount
	}
	return service.store.Transfer(ctx, fromAccountID, toAccountID, amount)
}
