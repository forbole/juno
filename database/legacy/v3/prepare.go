package v3

import (
	"fmt"

	"github.com/forbole/juno/v3/types/config"
)

// PrepareMigration prepares the database for the migration by renaming the old tables and creating the new ones
func (db *Migrator) PrepareMigration() error {
	err := db.renameCurrentTransactionsTable()
	if err != nil {
		return fmt.Errorf("error while altering transaction table: %s", err)
	}

	err = db.createNewTransactionsTable()
	if err != nil {
		return fmt.Errorf("error while creating transaction table: %s", err)
	}

	err = db.renameCurrentMessagesTable()
	if err != nil {
		return fmt.Errorf("error while altering message table: %s", err)
	}

	err = db.createNewMessagesTable()
	if err != nil {
		return fmt.Errorf("error while creating messaage table: %s", err)
	}

	err = db.migrateMessagesByAddressFunction()
	if err != nil {
		return fmt.Errorf("error while migrating the messages_by_address function: %s", err)
	}

	return nil
}

func (db *Migrator) renameCurrentTransactionsTable() error {
	stmt := `ALTER TABLE IF EXISTS transaction RENAME TO transaction_old;
ALTER INDEX IF EXISTS transaction_pkey RENAME TO transaction_old_pkey;
ALTER INDEX IF EXISTS transaction_hash_index RENAME TO transaction_old_hash_index;
ALTER INDEX IF EXISTS transaction_height_index RENAME TO transaction_old_height_index;
ALTER TABLE IF EXISTS transaction_old RENAME CONSTRAINT transaction_height_fkey TO transaction_old_height_fkey;`

	_, err := db.SQL.Exec(stmt)
	return err
}

func (db *Migrator) renameCurrentMessagesTable() error {
	stmt := `ALTER TABLE IF EXISTS message RENAME TO message_old;;
ALTER INDEX IF EXISTS message_involved_accounts_addresses RENAME TO message_old_involved_accounts_addresses;
ALTER INDEX IF EXISTS message_transaction_hash_index RENAME TO message_old_transaction_hash_index;
ALTER INDEX IF EXISTS message_type_index RENAME TO message_old_type_index;
ALTER TABLE IF EXISTS message_old RENAME CONSTRAINT message_transaction_hash_fkey TO message_old_transaction_hash_fkey;`

	_, err := db.SQL.Exec(stmt)
	return err
}

func (db *Migrator) createNewTransactionsTable() error {
	stmt := fmt.Sprintf(`
CREATE TABLE transaction
(
	hash         TEXT    NOT NULL,
	height       BIGINT  NOT NULL REFERENCES block (height),
	success      BOOLEAN NOT NULL,

	/* Body */
	messages     JSONB   NOT NULL DEFAULT '[]'::JSONB,
	memo         TEXT,
	signatures   TEXT[]  NOT NULL,

	/* AuthInfo */
	signer_infos JSONB   NOT NULL DEFAULT '[]'::JSONB,
	fee          JSONB   NOT NULL DEFAULT '{}'::JSONB,

	/* Tx response */
	gas_wanted   BIGINT           DEFAULT 0,
	gas_used     BIGINT           DEFAULT 0,
	raw_log      TEXT,
	logs         JSONB,

	/* PSQL partition */
	partition_id BIGINT NOT NULL,
	
	CONSTRAINT unique_tx UNIQUE (hash, partition_id)
) PARTITION BY LIST(partition_id);
CREATE INDEX transaction_hash_index ON transaction (hash);
CREATE INDEX transaction_height_index ON transaction (height);
CREATE INDEX transaction_partition_id_index ON transaction (partition_id);
GRANT ALL PRIVILEGES ON transaction TO "%s";
`,
		config.Cfg.Database.User)

	_, err := db.SQL.Exec(stmt)
	return err
}

func (db *Migrator) createNewMessagesTable() error {
	stmt := fmt.Sprintf(`
CREATE TABLE message
(
	transaction_hash            TEXT   NOT NULL,
	index                       BIGINT NOT NULL,
	type                        TEXT   NOT NULL,
	value                       JSONB  NOT NULL,
	involved_accounts_addresses TEXT[] NOT NULL,
	height                      BIGINT NOT NULL,

	/* PSQL partition */
	partition_id                BIGINT NOT NULL,
	
	FOREIGN KEY (transaction_hash, partition_id) REFERENCES transaction (hash, partition_id),  
	CONSTRAINT unique_message_per_tx UNIQUE (transaction_hash, index, partition_id)
) PARTITION BY LIST(partition_id);
CREATE INDEX message_transaction_hash_index ON message (transaction_hash);
CREATE INDEX message_type_index ON message (type);
CREATE INDEX message_involved_accounts_index ON message USING GIN(involved_accounts_addresses);
GRANT ALL PRIVILEGES ON message TO "%s";
`, config.Cfg.Database.User)

	_, err := db.SQL.Exec(stmt)
	return err
}

// migrateMessagesByAddressFunction migrates the messages_by_address function to use the new tables structures.
// This deletes the old function, which relied on a JOIN with the transaction table to get the message height,
// and creates a new function that does not require any JOIN since it takes the message height directly from
// the message table itself (as we've added this field with the new schema).
func (db *Migrator) migrateMessagesByAddressFunction() error {
	// Delete the old function
	_, err := db.SQL.Exec("DROP FUNCTION IF EXISTS messages_by_address(text[],text[],bigint,bigint);")
	if err != nil {
		return err
	}

	// Create the new function
	stmt := `
CREATE FUNCTION messages_by_address(
    addresses TEXT[],
    types TEXT[],
    "limit" BIGINT = 100,
    "offset" BIGINT = 0)
    RETURNS SETOF message AS
$$
SELECT * FROM message
WHERE (cardinality(types) = 0 OR type = ANY (types))
  AND addresses && involved_accounts_addresses
ORDER BY height DESC LIMIT "limit" OFFSET "offset"
$$ LANGUAGE sql STABLE;
`

	_, err = db.SQL.Exec(stmt)
	return err
}
