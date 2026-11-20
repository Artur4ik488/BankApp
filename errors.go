package main

import "errors"

var (
	ErrAccountNotFound   = errors.New("account not found")
	ErrAccountExists     = errors.New("account already exists")
	ErrInvalidAmount     = errors.New("invalid amount")
	ErrBalanceLimit      = errors.New("balance limit exceeded")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrSameAccount       = errors.New("cannot transfer to the same account")
	ErrInvalidOwner      = errors.New("invalid account owner")
)
