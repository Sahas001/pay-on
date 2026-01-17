-- migrations/000008_create_triggers.down.sql

DROP TRIGGER IF EXISTS update_wallets_updated_at ON wallets;
DROP TRIGGER IF EXISTS update_transactions_updated_at ON transactions;
DROP TRIGGER IF EXISTS update_peers_updated_at ON peers;
DROP TRIGGER IF EXISTS update_sync_logs_updated_at ON sync_logs;
DROP TRIGGER IF EXISTS trigger_audit_wallets ON wallets;
DROP TRIGGER IF EXISTS trigger_audit_transactions ON transactions;
DROP TRIGGER IF EXISTS trigger_update_peer_count ON transactions;

DROP FUNCTION IF EXISTS audit_wallet_changes() CASCADE;
DROP FUNCTION IF EXISTS audit_transaction_changes() CASCADE;
DROP FUNCTION IF EXISTS update_peer_transaction_count() CASCADE;
