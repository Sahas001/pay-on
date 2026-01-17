package database

import (
	"bytes"
	"context"
	"fmt"
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
		return err
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
