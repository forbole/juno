package migrate

import (
	"fmt"
)

func (db *MigrateDb) AlterTables() error {
	err := db.alterTxTable()
	if err != nil {
		return fmt.Errorf("error while altering transaction table: %s", err)
	}
	err = db.alterMsgTable()
	if err != nil {
		return fmt.Errorf("error while altering message table: %s", err)
	}
	return nil
}

func (db *MigrateDb) alterTxTable() error {
	stmt := `ALTER TABLE IF EXISTS transaction RENAME TO transaction_old;
ALTER INDEX IF EXISTS transaction_pkey RENAME TO transaction_old_pkey;
ALTER INDEX IF EXISTS transaction_hash_index RENAME TO transaction_old_hash_index;
ALTER INDEX IF EXISTS transaction_height_index RENAME TO transaction_old_height_index;
ALTER TABLE IF EXISTS transaction_old RENAME CONSTRAINT transaction_height_fkey TO transaction_old_height_fkey;`

	_, err := db.Sql.Exec(stmt)
	if err != nil {
		return err
	}

	fmt.Println(stmt)
	return nil
}

func (db *MigrateDb) alterMsgTable() error {
	stmt := `ALTER TABLE IF EXISTS message RENAME TO message_old;;
ALTER INDEX IF EXISTS message_involved_accounts_addresses RENAME TO message_old_involved_accounts_addresses;
ALTER INDEX IF EXISTS message_transaction_hash_index RENAME TO message_old_transaction_hash_index;
ALTER INDEX IF EXISTS message_type_index RENAME TO message_old_type_index;
ALTER TABLE IF EXISTS message_old RENAME CONSTRAINT message_transaction_hash_fkey TO message_old_transaction_hash_fkey;`

	_, err := db.Sql.Exec(stmt)
	if err != nil {
		return err
	}

	fmt.Println(stmt)
	return nil
}
