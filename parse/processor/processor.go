package processor

import (
	"github.com/angelorc/desmos-parser/db"
	"github.com/angelorc/desmos-parser/parse/client"
	"github.com/rs/zerolog/log"
)

type (
	// Queue is a simple type alias for a (buffered) channel of block heights.
	Queue chan int64

	// Worker defines a job consumer that is responsible for getting and
	// aggregating block and associated data and exporting it to a database.
	Worker struct {
		cp    client.ClientProxy
		queue Queue
		db    db.Database
	}
)

func NewQueue(size int) Queue {
	return make(chan int64, size)
}

func NewWorker(cp client.ClientProxy, q Queue, db db.Database) Worker {
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

	return db.HandleBlock(w.db, block, txs)
}
