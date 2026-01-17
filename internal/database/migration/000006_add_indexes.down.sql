-- migrations/000006_add_indexes.down.sql

DROP INDEX IF EXISTS idx_transactions_wallet_status;
DROP INDEX IF EXISTS idx_transactions_unsynced;
DROP INDEX IF EXISTS idx_wallets_active_sync;
DROP INDEX IF EXISTS idx_wallets_device;
DROP INDEX IF EXISTS idx_transactions_metadata;
DROP INDEX IF EXISTS idx_transactions_nonce;
DROP INDEX IF EXISTS idx_peers_connection_type;
