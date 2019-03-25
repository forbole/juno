package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

var configPath string

func main() {
	flag.StringVar(&configPath, "config", "", "Configuration file")
	flag.Parse()

	cfg := parseConfig(configPath)

	db, err := openDB(cfg)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to open database connection"))
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(errors.Wrap(err, "failed to ping database"))
	}

	rpc := newRPCClient(cfg.Node)

	execExportLoop(db, rpc)
}

func parseConfig(configPath string) config {
	if configPath == "" {
		log.Fatal("invalid configuration file")
	}

	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to read config"))
	}

	var cfg config
	if _, err := toml.Decode(string(configData), &cfg); err != nil {
		log.Fatal(errors.Wrap(err, "failed to decode config"))
	}

	return cfg
}

func execExportLoop(db *database, rpc rpcClient) {
	lastBlockHeight, err := db.lastBlockHeight()
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to get last block from database"))
	}

	latestBlockHeight, err := rpc.latestHeight()
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to get lastest block from RPC client"))
	}

	for i := int64(lastBlockHeight + 1); i <= latestBlockHeight; i++ {
		if i == 1 {
			// skip the first block since it has no pre-commits (violates table constraints)
			continue
		}

		if i%10 == 0 {
			log.Println("persisting block %d", i)
		}

		export(db, rpc, i)
	}
}

func export(db *database, rpc rpcClient, height int64) {
	block, err := rpc.block(height)
	if err != nil {
		log.Printf("failed to get block %d: %s\n", height, err)
		return
	}

	txs, err := rpc.txsFromBlock(block)
	if err != nil {
		log.Printf("failed to get transactions for block %d: %s\n", height, err)
		return
	}

	vals, err := rpc.validators(height)
	if err != nil {
		log.Printf("failed to get validators for block %d: %s\n", height, err)
		return
	}

	db.aggAndExport(block, txs, vals)
}
