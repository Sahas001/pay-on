-- internal/database/query/utils.sql

-- name: GetWalletDashboard :one
SELECT 
    w.id,
    w.name,
    w.phone_number,
    w.balance,
    w.last_synced_at,
    COUNT(DISTINCT CASE WHEN t.from_wallet_id = w.id OR t.to_wallet_id = w.id THEN t.id END) as total_transactions,
    COUNT(DISTINCT p.id) as total_peers,
    COUNT(DISTINCT CASE WHEN sl.status = 'pending' THEN sl.id END) as pending_syncs
FROM wallets w
LEFT JOIN transactions t ON (t.from_wallet_id = w.id OR t.to_wallet_id = w.id)
LEFT JOIN peers p ON p.wallet_id = w.id AND p.deleted_at IS NULL
LEFT JOIN sync_logs sl ON sl.wallet_id = w.id
WHERE w.id = $1 AND w.deleted_at IS NULL
GROUP BY w.id;

-- name: GetSystemStats :one
WITH wallet_stats AS (
    SELECT
        COUNT(*) as total_wallets,
        COUNT(*) FILTER (WHERE is_active) as active_wallets
    FROM wallets
),
transaction_stats AS (
    SELECT
        COUNT(*) as total_transactions,
        COALESCE(SUM(amount), 0) as total_volume
    FROM transactions
),
peer_stats AS (
    SELECT COUNT(*) as total_peers
    FROM peers
    WHERE deleted_at IS NULL
),
sync_stats AS (
    SELECT COUNT(*) as pending_syncs
    FROM sync_logs
    WHERE status = 'pending'
)
SELECT
    wallet_stats.total_wallets,
    wallet_stats.active_wallets,
    transaction_stats.total_transactions,
    transaction_stats.total_volume,
    peer_stats.total_peers,
    sync_stats.pending_syncs
FROM wallet_stats, transaction_stats, peer_stats, sync_stats;

-- name: SearchTransactions :many
SELECT 
    t.*,
    w_from.name as from_wallet_name,
    w_from.phone_number as from_wallet_phone,
    w_to.name as to_wallet_name,
    w_to.phone_number as to_wallet_phone
FROM transactions t
JOIN wallets w_from ON t.from_wallet_id = w_from.id
JOIN wallets w_to ON t.to_wallet_id = w_to.id
WHERE (
    w_from.name ILIKE '%' || $1 || '%'
    OR w_to.name ILIKE '%' || $1 || '%'
    OR w_from.phone_number LIKE '%' || $1 || '%'
    OR w_to.phone_number LIKE '%' || $1 || '%'
    OR t.metadata::text ILIKE '%' || $1 || '%'
)
ORDER BY t.transaction_at DESC
LIMIT $2 OFFSET $3;

-- name: GetWalletBalanceHistory :many
WITH balance_changes AS (
    SELECT 
        transaction_at as change_time,
        CASE 
            WHEN from_wallet_id = $1 THEN -amount
            WHEN to_wallet_id = $1 THEN amount
        END as change
    FROM transactions
    WHERE (from_wallet_id = $1 OR to_wallet_id = $1)
      AND status IN ('confirmed', 'settled')
    ORDER BY transaction_at
)
SELECT 
    change_time,
    change,
    SUM(change) OVER (ORDER BY change_time) as running_balance
FROM balance_changes
ORDER BY change_time DESC
LIMIT $2;
