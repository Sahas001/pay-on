-- migrations/000001_create_wallets_table. up.sql

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "btree_gist";

-- Create wallets table
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    public_key TEXT NOT NULL UNIQUE,
    private_key TEXT NOT NULL,
    balance DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    phone_number VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    pin_hash VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Metadata
    device_id VARCHAR(255),
    last_synced_at TIMESTAMP WITH TIME ZONE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT chk_positive_balance CHECK (balance >= 0),
    CONSTRAINT chk_valid_nepal_phone CHECK (
        phone_number ~ '^\+977[- ]?9[78][0-9]{8}$'
    ),
    CONSTRAINT chk_valid_name CHECK (
        LENGTH(TRIM(name)) >= 2
    )
);

-- Basic indexes
CREATE INDEX idx_wallets_phone ON wallets(phone_number) WHERE deleted_at IS NULL;
CREATE INDEX idx_wallets_public_key ON wallets(public_key);
CREATE INDEX idx_wallets_is_active ON wallets(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_wallets_created_at ON wallets(created_at DESC);

-- Comments
COMMENT ON TABLE wallets IS 'User wallet information with cryptographic keys and balance';
COMMENT ON COLUMN wallets.balance IS 'Current balance in NPR (Nepali Rupees)';
COMMENT ON COLUMN wallets.pin_hash IS 'Bcrypt hash of user PIN';
