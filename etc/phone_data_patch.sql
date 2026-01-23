/* Patch existing DB to match schema.prisma (safe-ish idempotent additions) */

-- Ensure transaction columns exist
ALTER TABLE transaction
  ADD COLUMN IF NOT EXISTS total_deposit BIGINT NOT NULL DEFAULT 0;

ALTER TABLE transaction
  ADD COLUMN IF NOT EXISTS total_profit BIGINT NOT NULL DEFAULT 0;

-- Ensure client columns exist
ALTER TABLE client
  ADD COLUMN IF NOT EXISTS created_at TIMESTAMP;

-- Ensure registration columns exist
ALTER TABLE registration
  ADD COLUMN IF NOT EXISTS created_at TIMESTAMP;

-- Ensure data columns exist
ALTER TABLE data
  ADD COLUMN IF NOT EXISTS created_at TIMESTAMP;

