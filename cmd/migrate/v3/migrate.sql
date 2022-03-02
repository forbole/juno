CREATE INDEX IF NOT EXISTS block_height_index ON block (height);

ALTER TABLE pre_commit ALTER COLUMN proposer_priority TYPE BIGINT;

ALTER TABLE transaction DROP CONSTRAINT IF EXISTS transaction_pkey;
ALTER TABLE transaction ADD COLUMN IF NOT EXISTS partition_id BIGINT NOT NULL DEFAULT 0;
ALTER TABLE transaction ADD CONSTRAINT unique_tx UNIQUE (hash, partition_id);
/* TODO: Check how to add the "PARTITION BY" here */

CREATE INDEX transaction_partition_id_index ON transaction (partition_id);

ALTER TABLE message DROP CONSTRAINT IF EXISTS message_transaction_hash_fkey;
ALTER TABLE message ADD COLUMN IF NOT EXISTS partition_id BIGINT NOT NULL DEFAULT 0;
ALTER TABLE message ADD FOREIGN KEY (transaction_hash, partition_id) REFERENCES transaction (hash, partition_id);
ALTER TABLE message ADD CONSTRAINT unique_message_per_tx UNIQUE (transaction_hash, index, partition_id);
/* TODO: Check how to add the "PARTITION BY" here */

CREATE INDEX IF NOT EXISTS message_type_index ON message (type);
CREATE INDEX IF NOT EXISTS message_involved_accounts_index ON message (involved_accounts_addresses);