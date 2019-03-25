package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

var (
	configPath string
	sync       bool
)

func main() {
	flag.StringVar(&configPath, "config", "", "Configuration file")
	flag.BoolVar(&sync, "sync", false, "Sync missing blocks from previous export")
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

	if sync {
		go syncMissing(db, rpc)
	}

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

// execExportLoop aggregates and exports blocks starting from the last persisted
// block height until the latest known block on the chain.
func execExportLoop(db *database, rpc rpcClient) {
	lastBlockHeight, err := db.lastBlockHeight()
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to get last block from database"))
	}

	latestBlockHeight, err := rpc.latestHeight()
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to get lastest block from RPC client"))
	}

	for i := lastBlockHeight + 1; i <= latestBlockHeight; i++ {
		if i == 1 {
			// skip the first block since it has no pre-commits (violates table constraints)
			continue
		}

		if i%10 == 0 {
			log.Printf("persisting block %d\n", i)
		}

		export(db, rpc, i)
	}
}

// syncMissing aggregates and exports missing blocks from a previous export in
// cases where the RPC client could fail or timeout but the export continued.
func syncMissing(db *database, rpc rpcClient) {
	lastBlockHeight, err := db.lastBlockHeight()
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to get last block from database"))
	}

	for i := int64(2); i < lastBlockHeight; i++ {
		ok, err := db.hasBlock(i)
		if !ok && err == nil {
			log.Printf("exporting missing block %d\n", i)
			export(db, rpc, i)
		}
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

	db.exportPreCommits(block, vals)
	db.exportBlock(block, txs)
}
