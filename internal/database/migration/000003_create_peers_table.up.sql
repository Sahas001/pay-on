-- migrations/000003_create_peers_table.up.sql

CREATE TABLE IF NOT EXISTS peers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL,
    peer_wallet_id UUID NOT NULL,
    name VARCHAR(100),
    public_key TEXT NOT NULL,
    
    -- Connection details
    ip_address INET,
    bt_address MACADDR,
    connection_type connection_type NOT NULL,
    
    -- Trust & reputation
    is_trusted BOOLEAN DEFAULT FALSE,
    transaction_count INTEGER DEFAULT 0,
    
    -- Activity tracking
    last_seen_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    first_seen_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Foreign keys
    CONSTRAINT fk_wallet FOREIGN KEY (wallet_id) 
        REFERENCES wallets(id) ON DELETE CASCADE,
    
    -- Constraints
    CONSTRAINT uq_wallet_peer UNIQUE (wallet_id, peer_wallet_id),
    CONSTRAINT chk_different_peer CHECK (wallet_id != peer_wallet_id),
    CONSTRAINT chk_positive_transaction_count CHECK (transaction_count >= 0)
);

-- Indexes
CREATE INDEX idx_peers_wallet ON peers(wallet_id, last_seen_at DESC);
CREATE INDEX idx_peers_peer_wallet ON peers(peer_wallet_id);
CREATE INDEX idx_peers_trusted ON peers(wallet_id, is_trusted) WHERE is_trusted = TRUE;
CREATE INDEX idx_peers_last_seen ON peers(last_seen_at DESC);

-- Comments
COMMENT ON TABLE peers IS 'Known peers for each wallet with connection history';
