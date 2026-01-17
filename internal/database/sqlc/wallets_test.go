package database

import (
	"context"
	"testing"
)

func TestWalletQueries(t *testing.T) {
	withTx(t, func(ctx context.Context, q *Queries) {
		wallet := createTestWallet(t, ctx, q)

		fetched, err := q.GetWalletByID(ctx, wallet.ID)
		if err != nil {
			t.Fatalf("get wallet by id: %v", err)
		}
		if fetched.ID != wallet.ID {
			t.Fatalf("expected wallet id %s, got %s", wallet.ID, fetched.ID)
		}

		newName := "Updated User"
		newDevice := "device-updated"
		updated, err := q.UpdateWallet(ctx, UpdateWalletParams{
			ID:       wallet.ID,
			Name:     &newName,
			DeviceID: &newDevice,
		})
		if err != nil {
			t.Fatalf("update wallet: %v", err)
		}
		if updated.Name != newName {
			t.Fatalf("expected name %q, got %q", newName, updated.Name)
		}

		updatedBalance, err := q.UpdateWalletBalance(ctx, UpdateWalletBalanceParams{
			ID:      wallet.ID,
			Balance: numericFromString(t, "250.50"),
		})
		if err != nil {
			t.Fatalf("update wallet balance: %v", err)
		}
		assertFloatApprox(t, numericToFloat64(t, updatedBalance.Balance), 250.50)

		incremented, err := q.IncrementWalletBalance(ctx, IncrementWalletBalanceParams{
			ID:      wallet.ID,
			Balance: numericFromString(t, "25.25"),
		})
		if err != nil {
			t.Fatalf("increment wallet balance: %v", err)
		}
		assertFloatApprox(t, numericToFloat64(t, incremented.Balance), 275.75)

		decremented, err := q.DecrementWalletBalance(ctx, DecrementWalletBalanceParams{
			ID:      wallet.ID,
			Balance: numericFromString(t, "75.00"),
		})
		if err != nil {
			t.Fatalf("decrement wallet balance: %v", err)
		}
		assertFloatApprox(t, numericToFloat64(t, decremented.Balance), 200.75)

		balance, err := q.GetWalletBalance(ctx, wallet.ID)
		if err != nil {
			t.Fatalf("get wallet balance: %v", err)
		}
		assertFloatApprox(t, numericToFloat64(t, balance), 200.75)

		if err := q.SoftDeleteWallet(ctx, wallet.ID); err != nil {
			t.Fatalf("soft delete wallet: %v", err)
		}
		if _, err := q.GetWalletByID(ctx, wallet.ID); err == nil {
			t.Fatalf("expected error after soft delete")
		}
	})
}
