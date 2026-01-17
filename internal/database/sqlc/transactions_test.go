package database

import (
	"context"
	"fmt"
	"testing"
)

func TestTransactionQueries(t *testing.T) {
	withTx(t, func(ctx context.Context, q *Queries) {
		fromWallet := createTestWallet(t, ctx, q)
		toWallet := createTestWallet(t, ctx, q)

		transaction := createTestTransaction(t, ctx, q, fromWallet.ID, toWallet.ID, "42.25", TransactionStatusConfirmed)

		fetched, err := q.GetTransactionByID(ctx, transaction.ID)
		if err != nil {
			t.Fatalf("get transaction by id: %v", err)
		}
		if fetched.ID != transaction.ID {
			t.Fatalf("expected transaction id %s, got %s", transaction.ID, fetched.ID)
		}

		exists, err := q.CheckNonceExists(ctx, CheckNonceExistsParams{
			FromWalletID: transaction.FromWalletID,
			Nonce:        transaction.Nonce,
		})
		if err != nil {
			t.Fatalf("check nonce exists: %v", err)
		}
		if !exists {
			t.Fatalf("expected nonce to exist")
		}

		rows, err := q.ListTransactionsByWallet(ctx, ListTransactionsByWalletParams{
			FromWalletID: fromWallet.ID,
			Limit:        10,
			Offset:       0,
		})
		if err != nil {
			t.Fatalf("list transactions by wallet: %v", err)
		}
		if len(rows) != 1 {
			t.Fatalf("expected 1 transaction, got %d", len(rows))
		}
		if direction := fmt.Sprint(rows[0].Direction); direction != "SENT" {
			t.Fatalf("expected direction SENT, got %s", direction)
		}

		withWallets, err := q.GetTransactionWithWallets(ctx, transaction.ID)
		if err != nil {
			t.Fatalf("get transaction with wallets: %v", err)
		}
		if withWallets.FromWalletName != fromWallet.Name {
			t.Fatalf("expected from wallet name %q, got %q", fromWallet.Name, withWallets.FromWalletName)
		}

		count, err := q.CountTransactionsByWallet(ctx, fromWallet.ID)
		if err != nil {
			t.Fatalf("count transactions by wallet: %v", err)
		}
		if count != 1 {
			t.Fatalf("expected transaction count 1, got %d", count)
		}
	})
}
