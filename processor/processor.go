package processor

import (
	"github.com/fissionlabsio/juno/client"
	"github.com/fissionlabsio/juno/db"
	"github.com/rs/zerolog/log"
)

type (
	// Queue is a simple type alias for a (buffered) channel of block heights.
	Queue chan int64

	// Worker defines a job consumer that is responsible for getting and
	// aggregating block and associated data and exporting it to a database.
	Worker struct {
		cp    client.ClientProxy
		db    *db.Database
		queue Queue
	}
)

func NewQueue(size int) Queue {
	return make(chan int64, size)
}

func NewWorker(db *db.Database, cp client.ClientProxy, q Queue) Worker {
	return Worker{cp, db, q}
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
	ok, err := w.db.HasBlock(height)
	if ok && err == nil {
		log.Debug().Int64("height", height).Msg("skipping already exported block")
		return nil
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

	vals, err := w.cp.Validators(block.Block.LastCommit.Height())
	if err != nil {
		log.Info().Err(err).Int64("height", height).Msg("failed to get validators for block")
		return err
	}

	if err := w.db.ExportPreCommits(block.Block.LastCommit, vals); err != nil {
		return err
	}

	return w.db.ExportBlock(block, txs, vals)
}
