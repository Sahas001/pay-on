-- migrations/000007_create_functions.down.sql

DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;
DROP FUNCTION IF EXISTS process_transaction(UUID, UUID, UUID, DECIMAL) CASCADE;
DROP FUNCTION IF EXISTS get_wallet_balance(UUID) CASCADE;
DROP FUNCTION IF EXISTS get_transaction_stats(UUID) CASCADE;
DROP FUNCTION IF EXISTS validate_transaction_nonce(UUID, BIGINT) CASCADE;
DROP FUNCTION IF EXISTS format_npr(DECIMAL) CASCADE;
DROP FUNCTION IF EXISTS cleanup_old_sync_logs(INTEGER) CASCADE;
DROP FUNCTION IF EXISTS auto_trust_peers() CASCADE;
