package database

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store provides transaction-safe access to queries.
type Store struct {
	*Queries
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		pool:    pool,
		Queries: New(pool),
	}
}

func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	q := store.Queries.WithTx(tx)
	if err := fn(q); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx error: %v, rb error: %v", err, rbErr)
		}
		return err
	}
	return tx.Commit(ctx)
}

// TransferTxParams contains the input parameters of the transfer transaction.
type TransferTxParams struct {
	FromWalletID   uuid.UUID
	ToWalletID     uuid.UUID
	Amount         pgtype.Numeric
	Currency       string
	Type           TransactionType
	Status         TransactionStatus
	Signature      string
	Nonce          int64
	ConnectionType NullConnectionType
	Description    *string
	Metadata       []byte
	TransactionAt  pgtype.Timestamptz
}

// TransferTxResult is the result of the transfer transaction.
type TransferTxResult struct {
	Transaction Transaction
	FromWallet  Wallet
	ToWallet    Wallet
}

// TransferTx performs a wallet-to-wallet transfer within a database transaction.
func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	if arg.FromWalletID == arg.ToWalletID {
		return result, fmt.Errorf("from and to wallet must be different")
	}

	if arg.Currency == "" {
		arg.Currency = "NPR"
	}
	if arg.Type == "" {
		arg.Type = TransactionTypeP2p
	}
	if arg.Status == "" {
		arg.Status = TransactionStatusPending
	}
	if len(arg.Metadata) == 0 {
		arg.Metadata = []byte(`{}`)
	}
	if arg.Nonce == 0 {
		arg.Nonce = time.Now().UnixNano()
	}
	if !arg.TransactionAt.Valid {
		arg.TransactionAt = pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	}

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Transaction, err = q.CreateTransaction(ctx, CreateTransactionParams(arg))
		if err != nil {
			return err
		}

		result.FromWallet, result.ToWallet, err = transferBalances(
			ctx,
			q,
			arg.FromWalletID,
			arg.ToWalletID,
			arg.Amount,
		)
		if err != nil {
			return err
		}

		if err := upsertTransferPeers(ctx, q, result.FromWallet, result.ToWallet, arg.ConnectionType); err != nil {
			return err
		}

		if err := incrementPeerCounts(ctx, q, result.FromWallet.ID, result.ToWallet.ID); err != nil {
			return err
		}

		return nil
	})

	return result, err
}

func transferBalances(
	ctx context.Context,
	q *Queries,
	fromWalletID uuid.UUID,
	toWalletID uuid.UUID,
	amount pgtype.Numeric,
) (fromWallet Wallet, toWallet Wallet, err error) {
	if bytes.Compare(fromWalletID[:], toWalletID[:]) < 0 {
		fromWallet, err = q.DecrementWalletBalance(ctx, DecrementWalletBalanceParams{
			ID:      fromWalletID,
			Balance: amount,
		})
		if err != nil {
			return fromWallet, toWallet, err
		}

		toWallet, err = q.IncrementWalletBalance(ctx, IncrementWalletBalanceParams{
			ID:      toWalletID,
			Balance: amount,
		})
		return fromWallet, toWallet, err
	}

	toWallet, err = q.IncrementWalletBalance(ctx, IncrementWalletBalanceParams{
		ID:      toWalletID,
		Balance: amount,
	})
	if err != nil {
		return fromWallet, toWallet, err
	}

	fromWallet, err = q.DecrementWalletBalance(ctx, DecrementWalletBalanceParams{
		ID:      fromWalletID,
		Balance: amount,
	})
	return fromWallet, toWallet, err
}

func upsertTransferPeers(ctx context.Context, q *Queries, fromWallet Wallet, toWallet Wallet, connType NullConnectionType) error {
	connection := connType
	if !connection.Valid {
		connection = NullConnectionType{ConnectionType: ConnectionTypeOnline, Valid: true}
	}

	trusted := false
	var btAddr net.HardwareAddr

	if _, err := q.UpsertPeer(ctx, UpsertPeerParams{
		WalletID:       fromWallet.ID,
		PeerWalletID:   toWallet.ID,
		Name:           &toWallet.Name,
		PublicKey:      toWallet.PublicKey,
		IpAddress:      nil,
		BtAddress:      btAddr,
		ConnectionType: connection.ConnectionType,
		IsTrusted:      &trusted,
	}); err != nil {
		return err
	}

	_, err := q.UpsertPeer(ctx, UpsertPeerParams{
		WalletID:       toWallet.ID,
		PeerWalletID:   fromWallet.ID,
		Name:           &fromWallet.Name,
		PublicKey:      fromWallet.PublicKey,
		IpAddress:      nil,
		BtAddress:      btAddr,
		ConnectionType: connection.ConnectionType,
		IsTrusted:      &trusted,
	})
	return err
}

func incrementPeerCounts(ctx context.Context, q *Queries, fromWalletID, toWalletID uuid.UUID) error {
	if err := q.IncrementPeerTransactionCount(ctx, IncrementPeerTransactionCountParams{
		WalletID:     fromWalletID,
		PeerWalletID: toWalletID,
	}); err != nil {
		return err
	}

	return q.IncrementPeerTransactionCount(ctx, IncrementPeerTransactionCountParams{
		WalletID:     toWalletID,
		PeerWalletID: fromWalletID,
	})
}
