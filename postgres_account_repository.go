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

var _ AccountRepository = (*PostgresAccountRepository)(nil)

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

func (repo *PostgresAccountRepository) ListAccounts(
	ctx context.Context,
) ([]Account, error) {
	const query = `
		SELECT id, owner, balance
		FROM accounts
		ORDER BY id
	`

	rows, err := repo.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := make([]Account, 0)
	for rows.Next() {
		var account Account
		if err := rows.Scan(&account.ID, &account.Owner, &account.Balance); err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return accounts, nil
}

func (repo *PostgresAccountRepository) UpdateAccount(
	ctx context.Context,
	account Account,
) error {
	const query = `
		UPDATE accounts
		SET owner = $1, balance = $2
		WHERE id = $3
	`

	result, err := repo.pool.Exec(
		ctx,
		query,
		account.Owner,
		account.Balance,
		account.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrAccountNotFound
	}

	return nil
}

func (repo *PostgresAccountRepository) UpdateAccounts(
	ctx context.Context,
	accountFrom, accountTo Account,
) error {
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	const query = `
		UPDATE accounts
		SET owner = $1, balance = $2
		WHERE id = $3
	`

	result1, err1 := tx.Exec(
		ctx,
		query,
		accountFrom.Owner,
		accountFrom.Balance,
		accountFrom.ID,
	)
	if err1 != nil {
		return err1
	}

	if result1.RowsAffected() == 0 {
		return ErrAccountNotFound
	}

	result2, err2 := tx.Exec(
		ctx,
		query,
		accountTo.Owner,
		accountTo.Balance,
		accountTo.ID,
	)
	if err2 != nil {
		return err2
	}

	if result2.RowsAffected() == 0 {
		return ErrAccountNotFound
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
