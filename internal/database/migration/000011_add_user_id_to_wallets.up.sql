-- migrations/000011_add_user_id_to_wallets.up.sql

ALTER TABLE wallets
    ADD COLUMN IF NOT EXISTS user_id UUID;

ALTER TABLE wallets
    ADD CONSTRAINT fk_wallet_user
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets(user_id);
