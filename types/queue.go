package types

import "github.com/forbole/juno/v5/interfaces"

var _ interfaces.BlockQueue = BlockQueue(nil)

// BlockQueue is a simple type alias for a (buffered) channel of block heights.
type BlockQueue chan interfaces.Block

func NewQueue(size int) BlockQueue {
	return make(chan interfaces.Block, size)
}

// Enqueue adds a block to the queue.
func (q BlockQueue) Enqueue(block interfaces.Block) {
	q <- block
}

// Listen returns the channel to listen for blocks.
func (q BlockQueue) Listen() <-chan interfaces.Block {
	return q
}
