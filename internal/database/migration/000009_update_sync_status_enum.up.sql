-- migrations/000009_update_sync_status_enum.up.sql

ALTER TABLE sync_logs ALTER COLUMN status DROP DEFAULT;

DROP INDEX IF EXISTS idx_sync_logs_status;
DROP INDEX IF EXISTS idx_sync_logs_wallet;

CREATE TYPE sync_status_new AS ENUM ('pending', 'confirmed', 'settling', 'settled', 'failed', 'conflict');

ALTER TABLE sync_logs
    ALTER COLUMN status TYPE text
    USING status::text;

ALTER TABLE sync_logs
    ALTER COLUMN status TYPE sync_status_new
    USING (CASE
        WHEN status = 'synced' THEN 'settled'
        ELSE status
    END)::sync_status_new;

DROP TYPE sync_status;
ALTER TYPE sync_status_new RENAME TO sync_status;

ALTER TABLE sync_logs ALTER COLUMN status SET DEFAULT 'pending';

CREATE INDEX idx_sync_logs_wallet ON sync_logs(wallet_id, status);
CREATE INDEX idx_sync_logs_status ON sync_logs(status, created_at) WHERE status = 'pending';
