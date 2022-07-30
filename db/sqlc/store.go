package db

import (
	"context"
	"database/sql"
	"fmt"
)

/*
This Store struct is needed because the Queries object contains the functions that
can run only one query on one particular table at a time. While for implmeneting transactions,
we need the ability to run multiple queries across multiple tables.
*/
type Store struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

type TransferCtxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferCtxResult struct {
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	Transfer    Transfer `json:"transfer"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

var txKey = struct{}{}

/*
This function handles the transaction of transfering money from one account to another
The operations involved are:
1. Create a transfer record
2. Add an entry in "from account"
3. Add an entry in "to account"
4. Update account balance with a single db transaction
*/
func (store *Store) TransferTxn(ctx context.Context, arg TransferCtxParams) (TransferCtxResult, error) {
	var result TransferCtxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		//1. Create a transfer record
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID:   arg.ToAccountID,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		//2. Add an entry in "from account"
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccID:  arg.FromAccountID,
			Amount: -arg.Amount,
		})
		if err != nil {
			return err
		}

		//3. Add an entry in "to account"
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccID:  arg.ToAccountID,
			Amount: arg.Amount,
		})
		if err != nil {
			return err
		}

		//4. Update account balance with a single db transaction
		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
		} else {
			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
		}

		return nil
	})

	return result, err
}

func addMoney(
	ctx context.Context,
	q *Queries,
	accountID1 int64,
	amount1 int64,
	accountID2 int64,
	amount2 int64,
) (account1 Account, account2 Account, err error) {
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID1,
		Amount: amount1,
	})
	if err != nil {
		return
	}
	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})
	return
}
