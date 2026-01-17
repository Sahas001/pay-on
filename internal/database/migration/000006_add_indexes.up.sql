-- migrations/000006_add_indexes.up.sql

-- Composite indexes for common queries
CREATE INDEX idx_transactions_wallet_status 
ON transactions(from_wallet_id, to_wallet_id, status, created_at DESC);

CREATE INDEX idx_transactions_unsynced 
ON transactions(status, created_at) 
WHERE status IN ('pending', 'confirmed');

CREATE INDEX idx_wallets_active_sync 
ON wallets(is_active, last_synced_at) 
WHERE is_active = TRUE;

CREATE INDEX idx_wallets_device 
ON wallets(device_id) 
WHERE device_id IS NOT NULL;

-- JSONB indexes
CREATE INDEX idx_transactions_metadata 
ON transactions USING GIN(metadata);

-- Nonce index for quick lookups
CREATE INDEX idx_transactions_nonce 
ON transactions(from_wallet_id, nonce);

-- Peer connection type index
CREATE INDEX idx_peers_connection_type 
ON peers(connection_type);
