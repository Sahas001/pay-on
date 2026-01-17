-- migrations/000007_create_functions.up.sql

-- Update timestamp trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Process transaction function
CREATE OR REPLACE FUNCTION process_transaction(
    p_transaction_id UUID,
    p_from_wallet_id UUID,
    p_to_wallet_id UUID,
    p_amount DECIMAL(15, 2)
)
RETURNS BOOLEAN AS $$
DECLARE
    v_from_balance DECIMAL(15, 2);
BEGIN
    -- Lock both wallet rows
    SELECT balance INTO v_from_balance
    FROM wallets
    WHERE id = p_from_wallet_id
    FOR UPDATE;
    
    -- Check sufficient balance
    IF v_from_balance < p_amount THEN
        RAISE EXCEPTION 'Insufficient balance';
    END IF;
    
    -- Deduct from sender
    UPDATE wallets
    SET balance = balance - p_amount
    WHERE id = p_from_wallet_id;
    
    -- Add to receiver
    UPDATE wallets
    SET balance = balance + p_amount
    WHERE id = p_to_wallet_id;
    
    -- Update transaction status
    UPDATE transactions
    SET status = 'confirmed',
        confirmed_at = NOW()
    WHERE id = p_transaction_id;
    
    RETURN TRUE;
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'Transaction failed: %', SQLERRM;
        RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

-- Get wallet balance function
CREATE OR REPLACE FUNCTION get_wallet_balance(p_wallet_id UUID)
RETURNS DECIMAL(15, 2) AS $$
DECLARE
    v_balance DECIMAL(15, 2);
BEGIN
    SELECT balance INTO v_balance
    FROM wallets
    WHERE id = p_wallet_id AND is_active = TRUE;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Wallet not found or inactive';
    END IF;
    
    RETURN v_balance;
END;
$$ LANGUAGE plpgsql;

-- Get transaction statistics function
CREATE OR REPLACE FUNCTION get_transaction_stats(p_wallet_id UUID)
RETURNS TABLE (
    total_sent DECIMAL(15, 2),
    total_received DECIMAL(15, 2),
    transaction_count BIGINT,
    net_flow DECIMAL(15, 2),
    avg_transaction DECIMAL(15, 2),
    last_transaction_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COALESCE(SUM(CASE WHEN t.from_wallet_id = p_wallet_id THEN t.amount ELSE 0 END), 0) AS total_sent,
        COALESCE(SUM(CASE WHEN t.to_wallet_id = p_wallet_id THEN t. amount ELSE 0 END), 0) AS total_received,
        COUNT(*)::BIGINT AS transaction_count,
        COALESCE(SUM(CASE WHEN t.to_wallet_id = p_wallet_id THEN t.amount ELSE -t.amount END), 0) AS net_flow,
        COALESCE(AVG(t.amount), 0) AS avg_transaction,
        MAX(t.transaction_at) AS last_transaction_at
    FROM transactions t
    WHERE (t.from_wallet_id = p_wallet_id OR t. to_wallet_id = p_wallet_id)
        AND t.status = 'confirmed';
END;
$$ LANGUAGE plpgsql;

-- Validate nonce function
CREATE OR REPLACE FUNCTION validate_transaction_nonce(
    p_wallet_id UUID,
    p_nonce BIGINT
)
RETURNS BOOLEAN AS $$
DECLARE
    v_exists BOOLEAN;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM transactions
        WHERE from_wallet_id = p_wallet_id AND nonce = p_nonce
    ) INTO v_exists;
    
    RETURN NOT v_exists;
END;
$$ LANGUAGE plpgsql;

-- Format NPR currency
CREATE OR REPLACE FUNCTION format_npr(amount DECIMAL(15, 2))
RETURNS TEXT AS $$
BEGIN
    RETURN 'रु ' || TO_CHAR(amount, 'FM999,999,999.00');
END;
$$ LANGUAGE plpgsql;

-- Cleanup old sync logs
CREATE OR REPLACE FUNCTION cleanup_old_sync_logs(days_to_keep INTEGER DEFAULT 90)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM sync_logs
    WHERE status = 'synced'
        AND updated_at < NOW() - (days_to_keep || ' days'):: INTERVAL;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Auto-trust frequent peers
CREATE OR REPLACE FUNCTION auto_trust_peers()
RETURNS VOID AS $$
BEGIN
    UPDATE peers
    SET is_trusted = TRUE
    WHERE transaction_count >= 5
        AND is_trusted = FALSE
        AND last_seen_at > NOW() - INTERVAL '30 days';
END;
$$ LANGUAGE plpgsql;
