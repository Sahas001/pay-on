-- internal/database/query/wallets.sql

-- name: CreateWallet :one
INSERT INTO wallets (
public_key,
private_key,
balance,
phone_number,
name,
pin_hash,
device_id
) VALUES (
	$1, $2, $3, $4, $5, $6, $7
	) RETURNING *;


-- name: GetWalletByID :one
SELECT * FROM wallets WHERE id = $1 AND deleted_at IS NULL;

-- name: GetWalletByPhoneNumber :one
SELECT * FROM wallets WHERE phone_number = $1 AND deleted_at IS NULL;

-- name: GetWalletByPublicKey :one
SELECT * FROM wallets WHERE public_key = $1 AND deleted_at IS NULL;

-- name: GetWalletByDeviceID :one
SELECT * FROM wallets WHERE device_id = $1 AND deleted_at IS NULL;

-- name: ListWallets :many
SELECT * FROM wallets WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: ListActiveWallets :many
SELECT * FROM wallets WHERE is_active = TRUE AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: UpdateWallet :one
UPDATE wallets
SET
name = COALESCE(sqlc.narg('name'), name),
phone_number = COALESCE(sqlc.narg('phone_number'), phone_number),
device_id = COALESCE(sqlc.narg('device_id'), device_id),
updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateWalletBalance :one
UPDATE wallets
SET
balance = $2,
updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: IncrementWalletBalance :one
UPDATE wallets
SET
balance = balance + $2,
updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DecrementWalletBalance :one
UPDATE wallets
SET
balance = balance - $2,
updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL AND balance >= $2
RETURNING *;

-- name: GetWalletBalance :one
SELECT balance FROM wallets
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateWalletPIN :exec
UPDATE wallets
SET
pin_hash = $2,
updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateWalletLastSync :exec
UPDATE wallets
SET
last_synced_at = NOW(),
updated_at = NOW()
WHERE id = $1;

-- name: DeactivateWallet :exec
UPDATE wallets
SET
is_active = FALSE,
updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: ActivateWallet :exec
UPDATE wallets
SET
is_active = TRUE,
updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteWallet :exec
UPDATE wallets
SET
deleted_at = NOW(),
is_active = FALSE,
updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: HardDeleteWallet :exec
DELETE FROM wallets
WHERE id = $1;

-- name: GetWalletWithBalance :one
SELECT id, name, phone_number, balance, is_active, created_at
FROM wallets
WHERE id = $1 AND deleted_at IS NULL;

-- name: SearchWalletsByName :many
SELECT id, name, phone_number, balance, is_active, created_at
FROM wallets
WHERE name ILIKE '%' || $1 || '%' 
	AND deleted_at IS NULL
	AND is_active = TRUE
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;


-- name: SearchWalletsByPhoneNumber :many
SELECT id, name, phone_number, balance, is_active, created_at
FROM wallets
WHERE phone_number ILIKE '%' || $1 || '%'
	AND deleted_at IS NULL
	AND is_active = TRUE
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountWallets :one
SELECT COUNT(*) AS count
FROM wallets
WHERE is_active = TRUE AND deleted_at IS NULL;


-- name: GetWalletsNeedingSync :many
SELECT * FROM wallets
WHERE (last_synced_at IS NULL OR last_synced_at < NOW() - INTERVAL '1 day')
	AND deleted_at IS NULL
ORDER BY last_synced_at ASC NULLS FIRST
LIMIT $1;
