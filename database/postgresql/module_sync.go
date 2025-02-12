package postgresql

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// SaveModuleSyncStatus implements database.Database
func (db *Database) SaveModuleSyncStatus(moduleName string, lastHeight int64, status string) error {
	stmt := `
		INSERT INTO module_sync_status (module_name, last_height, status) 
		VALUES ($1, $2, $3)
		ON CONFLICT (module_name) DO UPDATE 
		SET last_height = $2, 
		    status = $3,
		    updated_at = $4
	`

	_, err := db.SQL.Exec(stmt, moduleName, lastHeight, status, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("error while storing module sync status: %s", err)
	}

	return nil
}

// GetModuleSyncStatus implements database.Database
func (db *Database) GetModuleSyncStatus(moduleName string) (lastHeight int64, status string, err error) {
	stmt := `
		SELECT last_height, status 
		FROM module_sync_status 
		WHERE module_name = $1
	`

	err = db.SQL.QueryRow(stmt, moduleName).Scan(&lastHeight, &status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, "", fmt.Errorf("no sync status found for module %s", moduleName)
		}
		return 0, "", fmt.Errorf("error while getting module sync status: %s", err)
	}

	return lastHeight, status, nil
}

// UpdateModuleSyncStatus implements database.Database
func (db *Database) UpdateModuleSyncStatus(moduleName string, lastHeight int64, status string) error {
	stmt := `
		UPDATE module_sync_status 
		SET last_height = $1, 
		    status = $2,
		    updated_at = $3
		WHERE module_name = $4
	`

	result, err := db.SQL.Exec(stmt, lastHeight, status, time.Now().UTC(), moduleName)
	if err != nil {
		return fmt.Errorf("error while updating module sync status: %s", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error while getting affected rows: %s", err)
	}

	if rows == 0 {
		return fmt.Errorf("no module found with name %s", moduleName)
	}

	return nil
}
