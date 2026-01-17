-- internal/database/query/peers.sql

-- name: CreatePeer :one
INSERT INTO peers (
    wallet_id,
    peer_wallet_id,
    name,
    public_key,
    ip_address,
    bt_address,
    connection_type,
    is_trusted
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: UpsertPeer :one
INSERT INTO peers (
    wallet_id,
    peer_wallet_id,
    name,
    public_key,
    ip_address,
    bt_address,
    connection_type,
    is_trusted,
    last_seen_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, NOW()
)
ON CONFLICT (wallet_id, peer_wallet_id)
DO UPDATE SET
    name = EXCLUDED.name,
    ip_address = EXCLUDED.ip_address,
    bt_address = EXCLUDED.bt_address,
    connection_type = EXCLUDED.connection_type,
    last_seen_at = NOW(),
    updated_at = NOW(),
    deleted_at = NULL
RETURNING *;

-- name: GetPeerByID :one
SELECT * FROM peers
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetPeerByWalletAndPeerID :one
SELECT * FROM peers
WHERE wallet_id = $1 AND peer_wallet_id = $2 AND deleted_at IS NULL;

-- name: ListPeersByWallet :many
SELECT * FROM peers
WHERE wallet_id = $1 AND deleted_at IS NULL
ORDER BY last_seen_at DESC
LIMIT $2 OFFSET $3;

-- name: ListTrustedPeers :many
SELECT * FROM peers
WHERE wallet_id = $1 AND is_trusted = TRUE AND deleted_at IS NULL
ORDER BY last_seen_at DESC;

-- name: ListRecentPeers :many
SELECT * FROM peers
WHERE wallet_id = $1 
  AND last_seen_at >= NOW() - INTERVAL '7 days'
  AND deleted_at IS NULL
ORDER BY last_seen_at DESC
LIMIT $2;

-- name: ListPeersByConnectionType :many
SELECT * FROM peers
WHERE wallet_id = $1 AND connection_type = $2 AND deleted_at IS NULL
ORDER BY last_seen_at DESC
LIMIT $3;

-- name: UpdatePeerLastSeen :exec
UPDATE peers
SET last_seen_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: UpdatePeerInfo :one
UPDATE peers
SET 
    name = COALESCE(sqlc.narg('name'), name),
    ip_address = COALESCE(sqlc.narg('ip_address'), ip_address),
    bt_address = COALESCE(sqlc.narg('bt_address'), bt_address),
    connection_type = COALESCE(sqlc.narg('connection_type'), connection_type),
    last_seen_at = NOW(),
    updated_at = NOW()
WHERE wallet_id = $1 AND peer_wallet_id = $2
RETURNING *;

-- name: SetPeerTrusted :exec
UPDATE peers
SET is_trusted = $2, updated_at = NOW()
WHERE id = $1;

-- name: IncrementPeerTransactionCount :exec
UPDATE peers
SET 
    transaction_count = transaction_count + 1,
    last_seen_at = NOW(),
    updated_at = NOW()
WHERE wallet_id = $1 AND peer_wallet_id = $2;

-- name: GetTopPeersByVolume :many
SELECT 
    p.*,
    COALESCE(SUM(t.amount), 0) as total_volume
FROM peers p
LEFT JOIN transactions t ON (
    (t.from_wallet_id = p.wallet_id AND t.to_wallet_id = p.peer_wallet_id)
    OR
    (t.to_wallet_id = p.wallet_id AND t.from_wallet_id = p.peer_wallet_id)
)
WHERE p.wallet_id = $1 AND p.deleted_at IS NULL
GROUP BY p.id
ORDER BY total_volume DESC
LIMIT $2;

-- name: GetTopPeersByTransactionCount :many
SELECT * FROM peers
WHERE wallet_id = $1 AND deleted_at IS NULL
ORDER BY transaction_count DESC
LIMIT $2;

-- name: CountPeersByWallet :one
SELECT COUNT(*) FROM peers
WHERE wallet_id = $1 AND deleted_at IS NULL;

-- name: CountTrustedPeers :one
SELECT COUNT(*) FROM peers
WHERE wallet_id = $1 AND is_trusted = TRUE AND deleted_at IS NULL;

-- name: DeletePeer :exec
UPDATE peers
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: HardDeletePeer :exec
DELETE FROM peers
WHERE id = $1;

-- name: GetStalePeers :many
SELECT * FROM peers
WHERE last_seen_at < NOW() - INTERVAL '90 days'
  AND deleted_at IS NULL
ORDER BY last_seen_at ASC
LIMIT $1;

-- name: AutoTrustFrequentPeers :exec
UPDATE peers
SET is_trusted = TRUE, updated_at = NOW()
WHERE transaction_count >= $1
  AND is_trusted = FALSE
  AND last_seen_at > NOW() - INTERVAL '30 days'
  AND deleted_at IS NULL;
