package database

import (
	"context"
	"testing"
)

func TestWalletDashboard(t *testing.T) {
	withTx(t, func(ctx context.Context, q *Queries) {
		wallet := createTestWallet(t, ctx, q)
		peerWallet := createTestWallet(t, ctx, q)

		_ = createTestPeer(t, ctx, q, wallet.ID, peerWallet.ID)
		transaction := createTestTransaction(t, ctx, q, wallet.ID, peerWallet.ID, "12.00", TransactionStatusConfirmed)
		_ = createTestSyncLog(t, ctx, q, transaction.ID, wallet.ID, SyncStatusPending)

		dashboard, err := q.GetWalletDashboard(ctx, wallet.ID)
		if err != nil {
			t.Fatalf("get wallet dashboard: %v", err)
		}
		if dashboard.TotalTransactions != 1 {
			t.Fatalf("expected total transactions 1, got %d", dashboard.TotalTransactions)
		}
		if dashboard.TotalPeers != 1 {
			t.Fatalf("expected total peers 1, got %d", dashboard.TotalPeers)
		}
		if dashboard.PendingSyncs != 1 {
			t.Fatalf("expected pending syncs 1, got %d", dashboard.PendingSyncs)
		}
	})
}
