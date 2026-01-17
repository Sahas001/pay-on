-- migrations/000002_create_transactions_table.down.sql

DROP TABLE IF EXISTS transactions CASCADE;
DROP TYPE IF EXISTS transaction_status;
DROP TYPE IF EXISTS transaction_type;
DROP TYPE IF EXISTS connection_type;
