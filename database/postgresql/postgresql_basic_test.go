package postgresql_test

import (
	"database/sql"
)

func (s *DbTestSuite) TestRunTx() {
	// Check if errored transaction is rolled back
	err := s.database.RunTx(func(tx *sql.Tx) error {
		_, err := tx.Exec(`INSERT INTO validator(consensus_address, consensus_pubkey) VALUES ('addr', 'pub_key')`)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`INSERT INTO validator(consensus_address, consensus_pubkey) VALUES ('addr', 'pub_key')`)
		if err != nil {
			return err
		}

		return nil
	})
	s.Require().Error(err)

	var count int
	err = s.database.Sql.QueryRow(`SELECT COUNT(*) FROM validator`).Scan(&count)
	s.Require().Zero(count)

	// Check if valid transaction is committed
	err = s.database.RunTx(func(tx *sql.Tx) error {
		_, err = tx.Exec(`INSERT INTO validator(consensus_address, consensus_pubkey) VALUES ('addr1', 'pub_key1')`)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`INSERT INTO validator(consensus_address, consensus_pubkey) VALUES ('addr2', 'pub_key2')`)
		if err != nil {
			return err
		}

		return nil
	})
	s.Require().NoError(err)

	err = s.database.Sql.QueryRow(`SELECT COUNT(*) FROM validator`).Scan(&count)
	s.Require().Equal(2, count)
}
