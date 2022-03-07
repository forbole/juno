package migrate

import (
	"fmt"
)

func (db *MigrateDb) DropMessageByAddressFunc() error {
	fmt.Println("DROP FUNCTION messages_by_address()")
	_, err := db.Sqlx.Exec("DROP FUNCTION IF EXISTS messages_by_address(text[],text[],bigint,bigint);")
	if err != nil {
		return fmt.Errorf("error while dropping messages_by_address function: %s", err)
	}
	return nil
}
