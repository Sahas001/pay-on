-- migrations/000008_create_triggers.up.sql

-- Updated_at triggers for all tables
CREATE TRIGGER update_wallets_updated_at 
BEFORE UPDATE ON wallets
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_transactions_updated_at 
BEFORE UPDATE ON transactions
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_peers_updated_at 
BEFORE UPDATE ON peers
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sync_logs_updated_at 
BEFORE UPDATE ON sync_logs
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Audit triggers
CREATE OR REPLACE FUNCTION audit_wallet_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        INSERT INTO audit_logs (table_name, record_id, action, old_data)
        VALUES ('wallets', OLD.id, 'DELETE', row_to_json(OLD));
        RETURN OLD;
    ELSIF (TG_OP = 'UPDATE') THEN
        INSERT INTO audit_logs (table_name, record_id, action, old_data, new_data)
        VALUES ('wallets', NEW.id, 'UPDATE', row_to_json(OLD), row_to_json(NEW));
        RETURN NEW;
    ELSIF (TG_OP = 'INSERT') THEN
        INSERT INTO audit_logs (table_name, record_id, action, new_data)
        VALUES ('wallets', NEW.id, 'INSERT', row_to_json(NEW));
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_audit_wallets
AFTER INSERT OR UPDATE OR DELETE ON wallets
FOR EACH ROW EXECUTE FUNCTION audit_wallet_changes();

-- Transaction audit trigger
CREATE OR REPLACE FUNCTION audit_transaction_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        INSERT INTO audit_logs (table_name, record_id, action, old_data)
        VALUES ('transactions', OLD.id, 'DELETE', row_to_json(OLD));
        RETURN OLD;
    ELSIF (TG_OP = 'UPDATE') THEN
        INSERT INTO audit_logs (table_name, record_id, action, old_data, new_data)
        VALUES ('transactions', NEW.id, 'UPDATE', row_to_json(OLD), row_to_json(NEW));
        RETURN NEW;
    ELSIF (TG_OP = 'INSERT') THEN
        INSERT INTO audit_logs (table_name, record_id, action, new_data)
        VALUES ('transactions', NEW. id, 'INSERT', row_to_json(NEW));
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_audit_transactions
AFTER INSERT OR UPDATE OR DELETE ON transactions
FOR EACH ROW EXECUTE FUNCTION audit_transaction_changes();

-- Update peer transaction count trigger
CREATE OR REPLACE FUNCTION update_peer_transaction_count()
RETURNS TRIGGER AS $$
BEGIN
    -- Update peer count for sender
    UPDATE peers
    SET transaction_count = transaction_count + 1,
        last_seen_at = NOW()
    WHERE wallet_id = NEW.from_wallet_id 
        AND peer_wallet_id = NEW.to_wallet_id;
    
    -- Update peer count for receiver
    UPDATE peers
    SET transaction_count = transaction_count + 1,
        last_seen_at = NOW()
    WHERE wallet_id = NEW.to_wallet_id 
        AND peer_wallet_id = NEW.from_wallet_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_peer_count
AFTER INSERT ON transactions
FOR EACH ROW
WHEN (NEW.status = 'confirmed')
EXECUTE FUNCTION update_peer_transaction_count();
