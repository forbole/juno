package main

import (
	"log"

	"github.com/alexanderbez/juno/client"
	"github.com/alexanderbez/juno/db"
)

type (
	// queue is a simple type alias for a (buffered) channel of block heights.
	queue chan int64

	// worker defines a job consumer that is responsible for getting and
	// aggregating block and associated data and exporting it to a database.
	worker struct {
		client client.RPCClient
		db     *db.Database
		queue  queue
	}
)

func newQueue(size int) queue {
	return make(chan int64, size)
}

func newWorker(db *db.Database, rpc client.RPCClient, q queue) worker {
	return worker{rpc, db, q}
}

// start starts a worker by listening for new jobs (block heights) from the
// given worker queue. Any failed job is logged and re-enqueued.
func (w worker) start() {
	for i := range w.queue {
		if err := w.process(i); err != nil {
			// re-enqueue any failed job
			go func() {
				log.Printf("re-enqueueing failed block %d\n", i)
				w.queue <- i
			}()
		}
	}
}

// process defines the job consumer workflow. It will fetch a block for a given
// height and associated metadata and export it to a database. It returns an
// error if any export process fails.
func (w worker) process(height int64) error {
	ok, err := w.db.HasBlock(height)
	if ok && err == nil {
		log.Printf("skipping already exported block %d\n", height)
		return nil
	}

	block, err := w.client.Block(height)
	if err != nil {
		log.Printf("failed to get block %d: %s\n", height, err)
		return err
	}

	txs, err := w.client.TxsFromBlock(block)
	if err != nil {
		log.Printf("failed to get transactions for block %d: %s\n", height, err)
		return err
	}

	vals, err := w.client.Validators(block.Block.LastCommit.Height())
	if err != nil {
		log.Printf("failed to get validators for block %d: %s\n", height, err)
		return err
	}

	if err := w.db.ExportPreCommits(block.Block.LastCommit, vals); err != nil {
		return err
	}

	return w.db.ExportBlock(block, txs)
}
