package main

import (
	"flag"
	"log"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/desmos-labs/juno/client"
	"github.com/desmos-labs/juno/config"
	"github.com/desmos-labs/juno/db/postgresql"
	"github.com/pkg/errors"
)

var configPath string

func main() {
	flag.StringVar(&configPath, "config", "", "Configuration file")
	flag.Parse()

	cfg, err := config.ParseConfig(configPath)
	if err != nil {
		panic(err)
	}

	cp, err := client.New(cfg.RPCNode, cfg.ClientNode)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to start RPC client"))
	}

	defer cp.Stop() // nolint: errcheck

	codec := simapp.MakeCodec()
	db, err := postgresql.Builder(*cfg, codec)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to open database connection"))
	}

	defer db.Sql.Close()

	if err := db.Sql.Ping(); err != nil {
		log.Fatal(errors.Wrap(err, "failed to ping database"))
	}

	lastHeight, err := db.LastBlockHeight()
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to get latest block height"))
	}

	for i := int64(2); i < lastHeight; i++ {
		block, err := cp.Block(i)
		if err != nil {
			log.Printf("failed to get block %d: %s", i, err)
			continue
		}

		if i%10 == 0 {
			log.Printf("migrated block %d\n", i)
		}

		var id uint64
		err = db.Sql.QueryRow(
			`UPDATE block SET timestamp = $2 WHERE height = $1 RETURNING id;`,
			block.Block.Height, block.Block.Time,
		).Scan(&id)
		if err != nil {
			log.Printf("failed to persist block %d: %s", i, err)
		}
	}
}
