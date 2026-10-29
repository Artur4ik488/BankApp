package main

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresAccountRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresAccountRepository(
	pool *pgxpool.Pool,
) *PostgresAccountRepository {
	return &PostgresAccountRepository{
		pool: pool,
	}
}

func (repo *PostgresAccountRepository) CreateAccount(
	ctx context.Context,
	account Account,
) (Account, error) {

	const query = `
		INSERT INTO accounts (owner, balance)
		VALUES ($1, $2)
		RETURNING id, owner, balance 
	`

	row := repo.pool.QueryRow(
		ctx,
		query,
		account.Owner,
		account.Balance,
	)

	if err := row.Scan(
		&account.ID,
		&account.Owner,
		&account.Balance,
	); err != nil {
		return Account{}, err
	}

	return account, nil

}

func (repo *PostgresAccountRepository) GetAccount(
	ctx context.Context,
	id int,
) (Account, error) {
	var account Account

	const query = `
		SELECT id, owner, balance
		FROM accounts
		WHERE id = $1
	`

	err := repo.pool.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&account.ID,
		&account.Owner,
		&account.Balance,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, ErrAccountNotFound
	}

	if err != nil {
		return Account{}, err
	}

	return account, nil

}
