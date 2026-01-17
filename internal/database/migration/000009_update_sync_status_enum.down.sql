-- migrations/000009_update_sync_status_enum.down.sql

ALTER TABLE sync_logs ALTER COLUMN status DROP DEFAULT;

DROP INDEX IF EXISTS idx_sync_logs_status;
DROP INDEX IF EXISTS idx_sync_logs_wallet;

CREATE TYPE sync_status_old AS ENUM ('pending', 'synced', 'failed', 'conflict');

ALTER TABLE sync_logs
    ALTER COLUMN status TYPE text
    USING status::text;

ALTER TABLE sync_logs
    ALTER COLUMN status TYPE sync_status_old
    USING (CASE
        WHEN status IN ('confirmed', 'settling', 'settled') THEN 'synced'
        ELSE status
    END)::sync_status_old;

DROP TYPE sync_status;
ALTER TYPE sync_status_old RENAME TO sync_status;

ALTER TABLE sync_logs ALTER COLUMN status SET DEFAULT 'pending';

CREATE INDEX idx_sync_logs_wallet ON sync_logs(wallet_id, status);
CREATE INDEX idx_sync_logs_status ON sync_logs(status, created_at) WHERE status = 'pending';
