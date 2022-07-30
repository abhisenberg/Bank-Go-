package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransactionTx(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	amount := int64(10)
	n := 5

	errs := make(chan error)
	results := make(chan TransferCtxResult)

	//Make calls to create several transcations concurrently, and send their results in repective channels
	for i := 0; i < n; i++ {
		go func() {
			ctx := context.Background()
			result, err := store.TransferTxn(ctx, TransferCtxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})

			errs <- err
			results <- result
		}()
	}

	existed := make(map[int]bool)
	//Check the results received in the channels
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err) //Error shan't be present

		result := <-results
		require.NotEmpty(t, result) //Result shan't be empty

		//Check transfer record
		transfer := result.Transfer
		require.NotEmpty(t, transfer)
		require.Equal(t, account1.ID, transfer.FromAccountID)
		require.Equal(t, account2.ID, transfer.ToAccountID)
		require.Equal(t, transfer.Amount, amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err)

		//Check FromEntry record
		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		//Check ToEntry record
		toEntry := result.ToEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account2.ID, toEntry.AccID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, transfer.ID)
		require.NotZero(t, transfer.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		//Check accounts
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, fromAccount.ID, account1.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, toAccount.ID, account2.ID)

		//Check accounts balance
		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0) //amount, 2*amount, 3*amount ....

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	//Check balance in DB
	updatedAcc1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAcc2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance-int64(n)*amount, updatedAcc1.Balance)
	require.Equal(t, account2.Balance+int64(n)*amount, updatedAcc2.Balance)
}

func TestTransactionTxDeadlock(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	amount := int64(10)
	n := 10

	errs := make(chan error)

	//Make calls to create several transcations concurrently, and send their results in repective channels
	for i := 0; i < n; i++ {
		fromAccountID := account1.ID
		toAccountID := account2.ID

		if i%2 == 1 {
			fromAccountID = account2.ID
			toAccountID = account1.ID
		}

		go func() {
			ctx := context.Background()
			_, err := store.TransferTxn(ctx, TransferCtxParams{
				FromAccountID: fromAccountID,
				ToAccountID:   toAccountID,
				Amount:        amount,
			})

			errs <- err
		}()
	}

	//Check the results received in the channels
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err) //Error shan't be present
	}

	//Check balance in DB
	updatedAcc1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAcc2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance, updatedAcc1.Balance)
	require.Equal(t, account2.Balance, updatedAcc2.Balance)
}
