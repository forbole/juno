package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/fissionlabsio/juno/client"
	"github.com/fissionlabsio/juno/config"
	"github.com/fissionlabsio/juno/db"
	"github.com/pkg/errors"
)

var configPath string

func main() {
	flag.StringVar(&configPath, "config", "", "Configuration file")
	flag.Parse()

	cfg := config.ParseConfig(configPath)
	cp, err := client.New(cfg.RPCNode, cfg.ClientNode)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to start RPC client"))
	}

	defer cp.Stop() // nolint: errcheck

	db, err := db.OpenDB(cfg)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to open database connection"))
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
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

		for _, tmTx := range block.Block.Txs {
			txHash := fmt.Sprintf("%X", tmTx.Hash())

			tx, err := cp.Tx(txHash)
			if err != nil {
				log.Printf("failed to get tx %s: %s", txHash, err)
				continue
			}

			if i%10 == 0 {
				log.Printf("migrated transaction %s\n", txHash)
			}

			if _, err := db.SetTx(tx); err != nil {
				log.Printf("failed to persist transaction %s: %s", txHash, err)
			}
		}
	}
}
