-- migrations/000004_create_sync_logs_table.up.sql

-- Create ENUM type
CREATE TYPE sync_status AS ENUM ('pending', 'synced', 'failed', 'conflict');

CREATE TABLE IF NOT EXISTS sync_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL,
    wallet_id UUID NOT NULL,
    
    status sync_status NOT NULL DEFAULT 'pending',
    attempt_count INTEGER DEFAULT 0,
    last_attempt_at TIMESTAMP WITH TIME ZONE,
    
    -- Error tracking
    error_message TEXT,
    
    -- Conflict resolution
    conflict_data JSONB,
    resolved_at TIMESTAMP WITH TIME ZONE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Foreign keys
    CONSTRAINT fk_transaction FOREIGN KEY (transaction_id) 
        REFERENCES transactions(id) ON DELETE CASCADE,
    CONSTRAINT fk_sync_wallet FOREIGN KEY (wallet_id) 
        REFERENCES wallets(id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_sync_logs_transaction ON sync_logs(transaction_id);
CREATE INDEX idx_sync_logs_wallet ON sync_logs(wallet_id, status);
CREATE INDEX idx_sync_logs_status ON sync_logs(status, created_at) WHERE status = 'pending';
CREATE INDEX idx_sync_logs_conflict_data ON sync_logs USING GIN(conflict_data);

-- Comments
COMMENT ON TABLE sync_logs IS 'Synchronization logs for offline transactions';
