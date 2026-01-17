package database

import (
	"context"
	"testing"
)

func TestSyncLogQueries(t *testing.T) {
	withTx(t, func(ctx context.Context, q *Queries) {
		wallet := createTestWallet(t, ctx, q)
		peerWallet := createTestWallet(t, ctx, q)

		transaction := createTestTransaction(t, ctx, q, wallet.ID, peerWallet.ID, "15.00", TransactionStatusConfirmed)
		log := createTestSyncLog(t, ctx, q, transaction.ID, wallet.ID, SyncStatusPending)

		updated, err := q.UpdateSyncLogStatus(ctx, UpdateSyncLogStatusParams{
			ID:     log.ID,
			Status: SyncStatusFailed,
		})
		if err != nil {
			t.Fatalf("update sync log status: %v", err)
		}
		if updated.AttemptCount == nil || *updated.AttemptCount != 1 {
			t.Fatalf("expected attempt_count 1")
		}

		stats, err := q.GetSyncStats(ctx, wallet.ID)
		if err != nil {
			t.Fatalf("get sync stats: %v", err)
		}
		if stats.PendingCount != 0 || stats.FailedCount != 1 {
			t.Fatalf("unexpected sync stats: %+v", stats)
		}
	})
}
