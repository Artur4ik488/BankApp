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

func (repo *PostgresAccountRepository) Health(ctx context.Context) error {
	return repo.pool.Ping(ctx)
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

func (repo *PostgresAccountRepository) Deposit(
	ctx context.Context,
	accountID int,
	amount Money,
) (Account, error) {
	return repo.changeBalance(ctx, accountID, func(account *Account) error {
		return account.Deposit(amount)
	})
}

func (repo *PostgresAccountRepository) Withdraw(
	ctx context.Context,
	accountID int,
	amount Money,
) (Account, error) {
	return repo.changeBalance(ctx, accountID, func(account *Account) error {
		return account.Withdraw(amount)
	})
}

func (repo *PostgresAccountRepository) changeBalance(
	ctx context.Context,
	accountID int,
	change func(*Account) error,
) (Account, error) {
	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return Account{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	account, err := getAccountForUpdate(ctx, tx, accountID)
	if err != nil {
		return Account{}, err
	}
	if err := change(&account); err != nil {
		return Account{}, err
	}
	if err := saveAccountBalance(ctx, tx, account); err != nil {
		return Account{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Account{}, err
	}
	return account, nil
}

func (repo *PostgresAccountRepository) Transfer(
	ctx context.Context,
	fromAccountID, toAccountID int,
	amount Money,
) (Account, Account, error) {
	if fromAccountID == toAccountID {
		return Account{}, Account{}, ErrSameAccount
	}

	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return Account{}, Account{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	firstID, secondID := fromAccountID, toAccountID
	if firstID > secondID {
		firstID, secondID = secondID, firstID
	}

	firstAccount, err := getAccountForUpdate(ctx, tx, firstID)
	if err != nil {
		return Account{}, Account{}, err
	}
	secondAccount, err := getAccountForUpdate(ctx, tx, secondID)
	if err != nil {
		return Account{}, Account{}, err
	}

	accountFrom, accountTo := firstAccount, secondAccount
	if accountFrom.ID != fromAccountID {
		accountFrom, accountTo = secondAccount, firstAccount
	}

	if err := accountFrom.Withdraw(amount); err != nil {
		return Account{}, Account{}, err
	}
	if err := accountTo.Deposit(amount); err != nil {
		return Account{}, Account{}, err
	}
	if err := saveAccountBalance(ctx, tx, accountFrom); err != nil {
		return Account{}, Account{}, err
	}
	if err := saveAccountBalance(ctx, tx, accountTo); err != nil {
		return Account{}, Account{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Account{}, Account{}, err
	}
	return accountFrom, accountTo, nil
}

func getAccountForUpdate(ctx context.Context, tx pgx.Tx, accountID int) (Account, error) {
	var account Account
	const query = `
		SELECT id, owner, balance
		FROM accounts
		WHERE id = $1
		FOR UPDATE
	`
	err := tx.QueryRow(ctx, query, accountID).Scan(
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

func saveAccountBalance(ctx context.Context, tx pgx.Tx, account Account) error {
	const query = `
		UPDATE accounts
		SET balance = $1
		WHERE id = $2
	`
	result, err := tx.Exec(
		ctx,
		query,
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
