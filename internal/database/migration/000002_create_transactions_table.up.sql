-- migrations/000002_create_transactions_table.up.sql

-- Create ENUM types
CREATE TYPE transaction_status AS ENUM ('pending', 'confirmed', 'settling', 'settled', 'failed', 'rolled_back');
CREATE TYPE transaction_type AS ENUM ('p2p', 'deposit', 'withdraw');
CREATE TYPE connection_type AS ENUM ('lan', 'bluetooth', 'online');

-- Create transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_wallet_id UUID NOT NULL,
    to_wallet_id UUID NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'NPR',
    type transaction_type NOT NULL DEFAULT 'p2p',
    status transaction_status NOT NULL DEFAULT 'pending',
    
    -- Security
    signature TEXT NOT NULL,
    nonce BIGINT NOT NULL,
    
    -- Connection info
    connection_type connection_type,
    
    -- Metadata
    description TEXT,
    metadata JSONB DEFAULT '{}',
    
    -- Timestamps
    transaction_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    confirmed_at TIMESTAMP WITH TIME ZONE,
    synced_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Foreign keys
    CONSTRAINT fk_from_wallet FOREIGN KEY (from_wallet_id) 
        REFERENCES wallets(id) ON DELETE RESTRICT,
    CONSTRAINT fk_to_wallet FOREIGN KEY (to_wallet_id) 
        REFERENCES wallets(id) ON DELETE RESTRICT,
    
    -- Constraints
    CONSTRAINT chk_different_wallets CHECK (from_wallet_id != to_wallet_id),
    CONSTRAINT chk_positive_amount CHECK (amount > 0),
    CONSTRAINT chk_valid_amount CHECK (amount <= 1000000),
    CONSTRAINT uq_wallet_nonce UNIQUE (from_wallet_id, nonce)
);

-- Basic indexes
CREATE INDEX idx_transactions_from_wallet ON transactions(from_wallet_id, created_at DESC);
CREATE INDEX idx_transactions_to_wallet ON transactions(to_wallet_id, created_at DESC);
CREATE INDEX idx_transactions_status ON transactions(status, created_at DESC);
CREATE INDEX idx_transactions_transaction_at ON transactions(transaction_at DESC);

-- Comments
COMMENT ON TABLE transactions IS 'All payment transactions between wallets';
COMMENT ON COLUMN transactions.nonce IS 'Unique nonce per wallet for replay attack prevention';
COMMENT ON COLUMN transactions.signature IS 'ECDSA signature of transaction data';
