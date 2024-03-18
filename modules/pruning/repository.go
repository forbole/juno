package pruning

// PruningRepository represents a database that supports pruning properly
type PruningRepository interface {
	// Prune prunes the data for the given height, returning any error
	Prune(height int64) error

	// StoreLastPruned saves the last height at which the database was pruned
	StoreLastPruned(height int64) error

	// GetLastPruned returns the last height at which the database was pruned
	GetLastPruned() (int64, error)
}
