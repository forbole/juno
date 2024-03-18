package interfaces

import (
	"context"
	"time"
)

type Context interface {
	// Context represents the manager for the concurrent jobs
	context.Context

	// WorkerRepository returns the database used to operate blocks by workers
	WorkerRepository() WorkerRepository

	// BlockNode returns the client used to listen to blocks by enqueuers
	BlockNode() BlockNode

	// Modules returns a set of the extensional operations that should be executed inside a worker
	Modules() []Module

	// Logger returns the logger used to log the operations
	Logger() Logger
}

// Logger defines a function that takes an error and logs it.
type Logger interface {
	// SetLogLevel sets the log level
	SetLogLevel(level string) error

	// SetLogFormat sets the log format
	SetLogFormat(format string) error

	// Info logs the info message
	Info(msg string, keyvals ...interface{})

	// Debug logs the debug message
	Debug(msg string, keyvals ...interface{})

	// Error logs the error message
	Error(msg string, keyvals ...interface{})
}

type Block interface {
	// Height returns the height of the block.
	Height() int64

	// Hash returns the hash of the block.
	Hash() string

	// Timestamp returns the timestamp of the block.
	Timestamp() time.Time

	// Proposer returns the address of the block proposer.
	Proposer() string

	// Value returns the content of the block.
	Value() interface{}
}

type BlockQueue interface {
	// Enqueue enqueues a block into the queue.
	Enqueue(block Block)

	// Listen returns a channel to listen for blocks.
	Listen() <-chan Block
}

type BlockNode interface {
	// Block retrieves the block by the given height
	Block(height int64) (Block, error)

	// LatestBlock retrieves the latest block from the connected node.
	LatestBlock() (Block, error)

	// SubscribeBlocks returns a channel to subscribe to new blocks from the connected node.
	SubscribeBlocks() <-chan Block
}

type Enqueuer interface {
	// ListenAndEnqueueBlocks listens for available blocks and enqueues them into the queue.
	ListenAndEnqueueBlocks(ctx Context, queue BlockQueue)
}

type WorkerRepository interface {
	// SaveBlock saves a block into the database.
	SaveBlock(block Block) error

	// LatestBlock returns the latest block from the database.
	LatestBlock() (Block, error)
}

type Worker interface {
	// Start starts the worker to parse blocks from the queue.
	Start(ctx Context, queue BlockQueue)
}

type AdditionalOperator interface {
	// Register registers an additional operation module.
	Register(module AdditionalOperationsModule)

	// Start starts the additional operator.
	Start() error
}
