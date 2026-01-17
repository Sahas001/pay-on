package database

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestTransferTx(t *testing.T) {
	ctx := context.Background()
	store := NewStore(testPool)

	fromWallet := createTestWallet(t, ctx, store.Queries)
	toWallet := createTestWallet(t, ctx, store.Queries)
	var transactionID uuid.UUID

	defer func() {
		if transactionID != uuid.Nil {
			_, _ = testPool.Exec(ctx, "DELETE FROM transactions WHERE id = $1", transactionID)
		}
		_, _ = testPool.Exec(ctx, "DELETE FROM wallets WHERE id = $1 OR id = $2", fromWallet.ID, toWallet.ID)
	}()

	amount := numericFromString(t, "25.50")
	result, err := store.TransferTx(ctx, TransferTxParams{
		FromWalletID: fromWallet.ID,
		ToWalletID:   toWallet.ID,
		Amount:       amount,
		Signature:    "sig-test",
	})
	if err != nil {
		t.Fatalf("transfer tx: %v", err)
	}
	transactionID = result.Transaction.ID

	fromBalance := numericToFloat64(t, result.FromWallet.Balance)
	toBalance := numericToFloat64(t, result.ToWallet.Balance)

	assertFloatApprox(t, fromBalance, 74.50)
	assertFloatApprox(t, toBalance, 125.50)
}
