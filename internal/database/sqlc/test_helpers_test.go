package database

import (
	"context"
	"fmt"
	"math"
	"net"
	"net/netip"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var phoneSeq int64

func withTx(t *testing.T, fn func(ctx context.Context, q *Queries)) {
	t.Helper()

	ctx := context.Background()
	tx, err := testPool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	q := New(tx)
	fn(ctx, q)
}

func numericFromString(t *testing.T, value string) pgtype.Numeric {
	t.Helper()

	var n pgtype.Numeric
	if err := n.Scan(value); err != nil {
		t.Fatalf("scan numeric %q: %v", value, err)
	}
	return n
}

func numericToFloat64(t *testing.T, value pgtype.Numeric) float64 {
	t.Helper()

	f, err := value.Float64Value()
	if err != nil {
		t.Fatalf("numeric to float64: %v", err)
	}
	if !f.Valid {
		t.Fatalf("numeric value is not valid")
	}
	return f.Float64
}

func assertFloatApprox(t *testing.T, got, want float64) {
	t.Helper()

	if math.Abs(got-want) > 0.0001 {
		t.Fatalf("expected %.4f, got %.4f", want, got)
	}
}

func nextPhoneNumber() string {
	seq := atomic.AddInt64(&phoneSeq, 1)
	return fmt.Sprintf("+97798%08d", seq)
}

func pgUUIDFromUUID(id uuid.UUID) pgtype.UUID {
	var pgID pgtype.UUID
	copy(pgID.Bytes[:], id[:])
	pgID.Valid = true
	return pgID
}

func createTestWallet(t *testing.T, ctx context.Context, q *Queries) Wallet {
	t.Helper()

	deviceID := fmt.Sprintf("device-%s", uuid.NewString())
	name := fmt.Sprintf("User %d", atomic.AddInt64(&phoneSeq, 1))
	balance := numericFromString(t, "100.00")

	wallet, err := q.CreateWallet(ctx, CreateWalletParams{
		PublicKey:   "pub-" + uuid.NewString(),
		PrivateKey:  "priv-" + uuid.NewString(),
		Balance:     balance,
		PhoneNumber: nextPhoneNumber(),
		Name:        name,
		PinHash:     "pin-hash",
		DeviceID:    &deviceID,
	})
	if err != nil {
		t.Fatalf("create wallet: %v", err)
	}
	return wallet
}

func createTestTransaction(t *testing.T, ctx context.Context, q *Queries, fromID, toID uuid.UUID, amount string, status TransactionStatus) Transaction {
	t.Helper()

	description := "test transaction"
	metadata := []byte(`{"note":"test"}`)
	txTime := pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	connection := NullConnectionType{ConnectionType: ConnectionTypeOnline, Valid: true}

	transaction, err := q.CreateTransaction(ctx, CreateTransactionParams{
		FromWalletID:   fromID,
		ToWalletID:     toID,
		Amount:         numericFromString(t, amount),
		Currency:       "NPR",
		Type:           TransactionTypeP2p,
		Status:         status,
		Signature:      "sig-" + uuid.NewString(),
		Nonce:          time.Now().UnixNano(),
		ConnectionType: connection,
		Description:    &description,
		Metadata:       metadata,
		TransactionAt:  txTime,
	})
	if err != nil {
		t.Fatalf("create transaction: %v", err)
	}
	return transaction
}

func createTestPeer(t *testing.T, ctx context.Context, q *Queries, walletID, peerWalletID uuid.UUID) Peer {
	t.Helper()

	name := "Peer One"
	ip := netip.MustParseAddr("192.168.1.10")
	mac, err := net.ParseMAC("aa:bb:cc:dd:ee:ff")
	if err != nil {
		t.Fatalf("parse mac: %v", err)
	}
	trusted := true

	peer, err := q.CreatePeer(ctx, CreatePeerParams{
		WalletID:       walletID,
		PeerWalletID:   peerWalletID,
		Name:           &name,
		PublicKey:      "peer-pub-" + uuid.NewString(),
		IpAddress:      &ip,
		BtAddress:      mac,
		ConnectionType: ConnectionTypeLan,
		IsTrusted:      &trusted,
	})
	if err != nil {
		t.Fatalf("create peer: %v", err)
	}
	return peer
}

func createTestSyncLog(t *testing.T, ctx context.Context, q *Queries, transactionID, walletID uuid.UUID, status SyncStatus) SyncLog {
	t.Helper()

	log, err := q.CreateSyncLog(ctx, CreateSyncLogParams{
		TransactionID: transactionID,
		WalletID:      walletID,
		Status:        status,
	})
	if err != nil {
		t.Fatalf("create sync log: %v", err)
	}
	return log
}
