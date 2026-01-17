package database

import (
	"context"
	"net/netip"
	"testing"

	"github.com/google/uuid"
)

func TestAuditLogQueries(t *testing.T) {
	withTx(t, func(ctx context.Context, q *Queries) {
		recordID := uuid.New()
		changedBy := uuid.New()
		ip := netip.MustParseAddr("203.0.113.10")
		userAgent := "test-agent"

		log, err := q.CreateAuditLog(ctx, CreateAuditLogParams{
			TableName: "wallets",
			RecordID:  recordID,
			Action:    "UPDATE",
			OldData:   []byte(`{"balance":"10.00"}`),
			NewData:   []byte(`{"balance":"20.00"}`),
			ChangedBy: pgUUIDFromUUID(changedBy),
			IpAddress: &ip,
			UserAgent: &userAgent,
		})
		if err != nil {
			t.Fatalf("create audit log: %v", err)
		}

		fetched, err := q.GetAuditLogByID(ctx, log.ID)
		if err != nil {
			t.Fatalf("get audit log by id: %v", err)
		}
		if fetched.TableName != "wallets" {
			t.Fatalf("expected table name wallets, got %q", fetched.TableName)
		}

		count, err := q.CountAuditLogs(ctx)
		if err != nil {
			t.Fatalf("count audit logs: %v", err)
		}
		if count < 1 {
			t.Fatalf("expected audit log count >= 1")
		}
	})
}
