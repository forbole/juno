package migrate

import (
	"fmt"

	types "github.com/forbole/juno/v2/database/migrate/utils"
)

func (db *MigrateDb) SelectRows(limit int64, offset int64) ([]types.TransactionRow, error) {
	stmt := fmt.Sprintf("SELECT * FROM transaction_old ORDER BY height LIMIT %v OFFSET %v", limit, offset)
	var txRows []types.TransactionRow
	err := db.Sqlx.Select(&txRows, stmt)
	if err != nil {
		return nil, err
	}

	return txRows, nil
}
