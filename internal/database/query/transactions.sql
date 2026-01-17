-- internal/database/query/transactions.sql

-- name: CreateTransaction :one
INSERT INTO transactions (
    from_wallet_id,
    to_wallet_id,
    amount,
    currency,
    type,
    status,
    signature,
    nonce,
    connection_type,
    description,
    metadata,
    transaction_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: GetTransactionByID :one
SELECT * FROM transactions
WHERE id = $1;

-- name: GetTransactionWithWallets :one
SELECT 
    t.*,
    w_from.name as from_wallet_name,
    w_from.phone_number as from_wallet_phone,
    w_to.name as to_wallet_name,
    w_to.phone_number as to_wallet_phone
FROM transactions t
JOIN wallets w_from ON t.from_wallet_id = w_from.id
JOIN wallets w_to ON t.to_wallet_id = w_to.id
WHERE t.id = $1;

-- name: ListTransactionsByWallet :many
SELECT 
    t.*,
    CASE 
        WHEN t.from_wallet_id = $1 THEN 'SENT'
        WHEN t.to_wallet_id = $1 THEN 'RECEIVED'
    END as direction
FROM transactions t
WHERE (t.from_wallet_id = $1 OR t.to_wallet_id = $1)
ORDER BY t.transaction_at DESC
LIMIT $2 OFFSET $3;

-- name: ListSentTransactions :many
SELECT * FROM transactions
WHERE from_wallet_id = $1
ORDER BY transaction_at DESC
LIMIT $2 OFFSET $3;

-- name: ListReceivedTransactions :many
SELECT * FROM transactions
WHERE to_wallet_id = $1
ORDER BY transaction_at DESC
LIMIT $2 OFFSET $3;

-- name: ListTransactionsByStatus :many
SELECT * FROM transactions
WHERE status = $1
ORDER BY created_at ASC
LIMIT $2 OFFSET $3;

-- name: ListPendingTransactions :many
SELECT * FROM transactions
WHERE status IN ('pending', 'confirmed')
  AND (from_wallet_id = $1 OR to_wallet_id = $1)
ORDER BY created_at ASC
LIMIT $2;

-- name: ListUnsyncedTransactions :many
SELECT * FROM transactions
WHERE status IN ('pending', 'confirmed')
ORDER BY created_at ASC
LIMIT $1 OFFSET $2;

-- name: GetTransactionsByDateRange :many
SELECT * FROM transactions
WHERE (from_wallet_id = $1 OR to_wallet_id = $1)
  AND transaction_at BETWEEN $2 AND $3
ORDER BY transaction_at DESC;

-- name: UpdateTransactionStatus :one
UPDATE transactions
SET 
    status = $2,
    confirmed_at = CASE WHEN $2 = 'confirmed' THEN NOW() ELSE confirmed_at END,
    synced_at = CASE WHEN $2 = 'synced' THEN NOW() ELSE synced_at END,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ConfirmTransaction :one
UPDATE transactions
SET 
    status = 'confirmed',
    confirmed_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SettingTransaction :one
UPDATE transactions
SET
    status = 'setting',
    updated_at = NOW()
WHERE id = $1 AND status = 'confirmed'
RETURNING *;

-- name: SettledTransaction :one
UPDATE transactions
SET
    status = 'settled',
    synced_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND status IN ('confirmed', 'settling')
RETURNING *;


-- name: MarkTransactionSettled :one
UPDATE transactions
SET 
    status = 'settled',
    synced_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: FailTransaction :exec
UPDATE transactions
SET 
    status = 'failed',
    updated_at = NOW()
WHERE id = $1;

-- name: CheckNonceExists :one
SELECT EXISTS(
    SELECT 1 FROM transactions
    WHERE from_wallet_id = $1 AND nonce = $2
);

-- name: GetTransactionStats :one
SELECT 
    COALESCE(SUM(CASE WHEN from_wallet_id = $1 THEN amount ELSE 0 END), 0) as total_sent,
    COALESCE(SUM(CASE WHEN to_wallet_id = $1 THEN amount ELSE 0 END), 0) as total_received,
    COUNT(*) as transaction_count,
    COALESCE(SUM(CASE WHEN to_wallet_id = $1 THEN amount ELSE -amount END), 0) as net_flow,
    COALESCE(AVG(amount), 0) as avg_transaction,
    MAX(transaction_at) as last_transaction_at
FROM transactions
WHERE (from_wallet_id = $1 OR to_wallet_id = $1)
  AND status IN ('confirmed', 'settled');

-- name: GetDailyTransactionSummary :many
SELECT 
    DATE(transaction_at) as date,
    COUNT(*) as transaction_count,
    SUM(CASE WHEN from_wallet_id = $1 THEN amount ELSE 0 END) as total_sent,
    SUM(CASE WHEN to_wallet_id = $1 THEN amount ELSE 0 END) as total_received,
    SUM(CASE WHEN to_wallet_id = $1 THEN amount ELSE -amount END) as net_amount
FROM transactions
WHERE (from_wallet_id = $1 OR to_wallet_id = $1)
  AND status IN ('confirmed', 'settled')
  AND transaction_at >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY DATE(transaction_at)
ORDER BY date DESC;

-- name: CountTransactionsByWallet :one
SELECT COUNT(*) FROM transactions
WHERE (from_wallet_id = $1 OR to_wallet_id = $1);

-- name: CountPendingTransactions :one
SELECT COUNT(*) FROM transactions
WHERE status = 'pending';

-- name: GetTransactionsByMetadata :many
SELECT * FROM transactions
WHERE metadata @> $1:: jsonb
ORDER BY transaction_at DESC
LIMIT $2 OFFSET $3;

-- name: GetRecentTransactions :many
SELECT 
    t.*,
    w_from.name as from_wallet_name,
    w_to.name as to_wallet_name
FROM transactions t
JOIN wallets w_from ON t.from_wallet_id = w_from. id
JOIN wallets w_to ON t.to_wallet_id = w_to.id
WHERE t.transaction_at >= NOW() - INTERVAL '24 hours'
  AND t. status IN ('confirmed', 'settled')
ORDER BY t.transaction_at DESC
LIMIT $1;

-- name: GetTransactionsByConnectionType :many
SELECT * FROM transactions
WHERE connection_type = $1
  AND transaction_at >= $2
ORDER BY transaction_at DESC
LIMIT $3;

-- name: GetLargeTransactions :many
SELECT * FROM transactions
WHERE amount >= $1
  AND status IN ('confirmed', 'settled')
ORDER BY amount DESC, transaction_at DESC
LIMIT $2;
