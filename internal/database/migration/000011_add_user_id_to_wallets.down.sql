-- migrations/000011_add_user_id_to_wallets.down.sql

ALTER TABLE wallets DROP CONSTRAINT IF EXISTS fk_wallet_user;
DROP INDEX IF EXISTS idx_wallets_user_id;
ALTER TABLE wallets DROP COLUMN IF EXISTS user_id;
