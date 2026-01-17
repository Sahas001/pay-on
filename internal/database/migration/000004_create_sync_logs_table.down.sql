-- migrations/000004_create_sync_logs_table.down.sql

DROP TABLE IF EXISTS sync_logs CASCADE;
DROP TYPE IF EXISTS sync_status;
