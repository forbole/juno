package worker

import (
	"fmt"

	"github.com/angelorc/desmos-parser/db"
	"github.com/angelorc/desmos-parser/parse/client"
	"github.com/angelorc/desmos-parser/types"
	"github.com/rs/zerolog/log"
)

// Worker defines a job consumer that is responsible for getting and
// aggregating block and associated data and exporting it to a database.
type Worker struct {
	cp    client.ClientProxy
	queue types.Queue
	db    db.Database
}

func NewWorker(cp client.ClientProxy, q types.Queue, db db.Database) Worker {
	return Worker{cp, q, db}
}

// Start starts a worker by listening for new jobs (block heights) from the
// given worker queue. Any failed job is logged and re-enqueued.
func (w Worker) Start() {
	for i := range w.queue {
		log.Info().Int64("height", i).Msg("processing block")

		if err := w.process(i); err != nil {
			// re-enqueue any failed job
			// TODO: Implement exponential backoff or max retries for a block height.
			go func() {
				log.Info().Int64("height", i).Msg("re-enqueueing failed block")
				w.queue <- i
			}()
		}
	}
}

// process defines the job consumer workflow. It will fetch a block for a given
// height and associated metadata and export it to a database. It returns an
// error if any export process fails.
func (w Worker) process(height int64) error {
	if exists := w.db.HasBlock(height); exists {
		log.Debug().Int64("height", height).Msg("skipping already exported block with mongodb")
		return nil
	}

	if height == 1 {
		log.Info().Msg("Parse genesis")

		/*if err := w.db.CreateIndexes(); err != nil {
			log.Info().Err(err).Int64("height", height).Msg("error creating index")
		}*/

		/*genesis, err := w.cp.Genesis()
		if err != nil {
			log.Info().Err(err).Int64("height", height).Msg("failed to get genesis")
		}

		return w.db.ExportGenesis(genesis)*/
	}

	block, err := w.cp.Block(height)
	if err != nil {
		log.Info().Err(err).Int64("height", height).Msg("failed to get block")
		return err
	}

	txs, err := w.cp.Txs(block)
	if err != nil {
		log.Info().Err(err).Int64("height", height).Msg("failed to get transactions for block")
		return err
	}

	/*blockResults, err := w.cp.BlockResults(height)
	if err != nil {
		log.Info().Err(err).Int64("height", height).Msg("failed to get block results")
		return err
	}

	vals, err := w.cp.Validators(block.Block.LastCommit.Height())
	if err != nil {
		log.Info().Err(err).Int64("height", height).Msg("failed to get validators for block")
		return err
	}*/

	// Save the block
	if err := db.SaveBlock(w.db, block, txs); err != nil {
		return err
	}

	// Handle all the transactions inside the block
	for _, tx := range txs {
		// Convert the transaction to a more easy-to-handle type
		txData, err := types.NewTx(tx)
		if err != nil {
			return fmt.Errorf("error handleTx")
		}

		// Save the transaction itself
		if err := db.SaveTx(w.db, *txData); err != nil {
			return err
		}

		// Handle all the messages contained inside the transaction
		for i, msg := range txData.Messages {
			if err := w.db.SaveMsg(*txData, i, msg); err != nil {
				return err
			}
		}
	}

	return nil
}
