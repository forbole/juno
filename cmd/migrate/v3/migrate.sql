CREATE INDEX IF NOT EXISTS block_height_index ON block (height);

ALTER TABLE pre_commit ALTER COLUMN proposer_priority TYPE BIGINT;
