package postgresql_test

import (
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	databaseconfig "github.com/forbole/juno/v5/database/config"
	"github.com/forbole/juno/v5/database/postgresql"
	postgres "github.com/forbole/juno/v5/database/postgresql"
	"github.com/forbole/juno/v5/logging"
	"github.com/forbole/juno/v5/types/params"
)

func TestDatabaseTestSuite(t *testing.T) {
	suite.Run(t, new(DbTestSuite))
}

type DbTestSuite struct {
	suite.Suite

	database *postgres.Database
}

func (suite *DbTestSuite) SetupTest() {
	// Create the codec
	codec := params.MakeTestEncodingConfig()

	// Build the database config
	dbCfg := databaseconfig.DefaultDatabaseConfig().
		WithURL("postgres://bdjuno:password@localhost:6433/bdjuno?sslmode=disable&search_path=public")

	// Build the database
	db, err := postgres.Builder(postgresql.NewContext(dbCfg, codec, logging.DefaultLogger()))
	suite.Require().NoError(err)

	// Delete the public schema
	_, err = db.SQL.Exec(`DROP SCHEMA public CASCADE;`)
	suite.Require().NoError(err)

	// Re-create the schema
	_, err = db.SQL.Exec(`CREATE SCHEMA public;`)
	suite.Require().NoError(err)

	dirPath := path.Join(".")
	dir, err := ioutil.ReadDir(dirPath)
	suite.Require().NoError(err)

	for _, fileInfo := range dir {
		if !strings.Contains(fileInfo.Name(), ".sql") {
			continue
		}

		file, err := ioutil.ReadFile(filepath.Join(dirPath, fileInfo.Name()))
		suite.Require().NoError(err)

		commentsRegExp := regexp.MustCompile(`/\*.*\*/`)
		requests := strings.Split(string(file), ";")
		for _, request := range requests {
			_, err := db.SQL.Exec(commentsRegExp.ReplaceAllString(request, ""))
			suite.Require().NoError(err)
		}
	}

	suite.database = db
}
