-- internal/database/query/audit_logs.sql

-- name: CreateAuditLog :one
INSERT INTO audit_logs (
    table_name,
    record_id,
    action,
    old_data,
    new_data,
    changed_by,
    ip_address,
    user_agent
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetAuditLogByID :one
SELECT * FROM audit_logs
WHERE id = $1;

-- name: ListAuditLogs :many
SELECT * FROM audit_logs
ORDER BY changed_at DESC
LIMIT $1 OFFSET $2;

-- name: ListAuditLogsByTable :many
SELECT * FROM audit_logs
WHERE table_name = $1
ORDER BY changed_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAuditLogsByRecord :many
SELECT * FROM audit_logs
WHERE table_name = $1 AND record_id = $2
ORDER BY changed_at DESC;

-- name: ListAuditLogsByAction :many
SELECT * FROM audit_logs
WHERE action = $1
ORDER BY changed_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAuditLogsByUser :many
SELECT * FROM audit_logs
WHERE changed_by = $1
ORDER BY changed_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAuditLogsByDateRange :many
SELECT * FROM audit_logs
WHERE changed_at BETWEEN $1 AND $2
ORDER BY changed_at DESC
LIMIT $3 OFFSET $4;

-- name: ListAuditLogsByIP :many
SELECT * FROM audit_logs
WHERE ip_address = $1
ORDER BY changed_at DESC
LIMIT $2 OFFSET $3;

-- name: GetRecordHistory :many
SELECT * FROM audit_logs
WHERE table_name = $1 AND record_id = $2
ORDER BY changed_at ASC;

-- name: GetRecentAuditLogs :many
SELECT * FROM audit_logs
WHERE changed_at >= NOW() - INTERVAL '24 hours'
ORDER BY changed_at DESC
LIMIT $1;

-- name: CountAuditLogs :one
SELECT COUNT(*) FROM audit_logs;

-- name: CountAuditLogsByTable :one
SELECT COUNT(*) FROM audit_logs
WHERE table_name = $1;

-- name: DeleteOldAuditLogs :exec
DELETE FROM audit_logs
WHERE changed_at < NOW() - ($1 || ' days')::INTERVAL;

-- name: GetBalanceHistory :many
SELECT 
    changed_at,
    old_data->>'balance' as old_balance,
    new_data->>'balance' as new_balance
FROM audit_logs
WHERE table_name = 'wallets'
  AND record_id = $1
  AND action = 'UPDATE'
  AND (old_data->>'balance' IS DISTINCT FROM new_data->>'balance')
ORDER BY changed_at DESC
LIMIT $2;
