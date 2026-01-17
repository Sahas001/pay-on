-- internal/database/query/sync_logs.sql

-- name: CreateSyncLog :one
INSERT INTO sync_logs (
    transaction_id,
    wallet_id,
    status
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetSyncLogByID :one
SELECT * FROM sync_logs
WHERE id = $1;

-- name: GetSyncLogsByTransaction :many
SELECT * FROM sync_logs
WHERE transaction_id = $1
ORDER BY created_at DESC;

-- name: GetSyncLogsByWallet :many
SELECT * FROM sync_logs
WHERE wallet_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListPendingSyncs :many
SELECT 
    sl.*,
    t.amount,
    t.type as transaction_type,
    t.created_at as transaction_created_at
FROM sync_logs sl
JOIN transactions t ON sl.transaction_id = t.id
WHERE sl.status = 'pending'
  AND sl.wallet_id = $1
ORDER BY sl.created_at ASC
LIMIT $2;

-- name: ListAllPendingSyncs :many
SELECT * FROM sync_logs
WHERE status = 'pending'
ORDER BY created_at ASC
LIMIT $1 OFFSET $2;

-- name: ListFailedSyncs :many
SELECT * FROM sync_logs
WHERE status = 'failed'
  AND wallet_id = $1
ORDER BY last_attempt_at DESC
LIMIT $2;

-- name: ListConflictedSyncs :many
SELECT * FROM sync_logs
WHERE status = 'conflict'
  AND wallet_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: UpdateSyncLogStatus :one
UPDATE sync_logs
SET 
    status = $2,
    last_attempt_at = NOW(),
    attempt_count = attempt_count + 1,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: MarkSettleSuccessful :one
UPDATE sync_logs
SET 
    status = 'settled',
    last_attempt_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: MarkSettleFailed :one
UPDATE sync_logs
SET 
    status = 'failed',
    last_attempt_at = NOW(),
    attempt_count = attempt_count + 1,
    error_message = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: MarkSettleConflict :one
UPDATE sync_logs
SET 
    status = 'conflict',
    last_attempt_at = NOW(),
    attempt_count = attempt_count + 1,
    conflict_data = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ResolveSyncConflict :one
UPDATE sync_logs
SET 
    status = 'settled',
    resolved_at = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetSyncStats :one
SELECT 
    COUNT(*) FILTER (WHERE status = 'pending') as pending_count,
    COUNT(*) FILTER (WHERE status = 'settled') as synced_count,
    COUNT(*) FILTER (WHERE status = 'failed') as failed_count,
    COUNT(*) FILTER (WHERE status = 'conflict') as conflict_count,
    AVG(attempt_count) as avg_attempts
FROM sync_logs
WHERE wallet_id = $1;

-- name: GetSyncsNeedingRetry :many
SELECT * FROM sync_logs
WHERE status = 'failed'
  AND attempt_count < $1
  AND (last_attempt_at IS NULL OR last_attempt_at < NOW() - INTERVAL '5 minutes')
ORDER BY attempt_count ASC, created_at ASC
LIMIT $2;

-- name: DeleteOldSyncLogs :exec
DELETE FROM sync_logs
WHERE status = 'settled'
  AND updated_at < NOW() - ($1 || ' days')::INTERVAL;

-- name: CountSyncLogsByStatus :one
SELECT COUNT(*) FROM sync_logs
WHERE status = $1;
