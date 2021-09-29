package node

const (
	LocalKeeper  = "local"
	RemoteKeeper = "remote"
)

// Source represents a generic source that allows to read the data of a specific SDK module
type Source interface {

	// Type returns whether the keeper is a LocalKeeper or a RemoteKeeper
	Type() string
}
