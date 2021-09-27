package node

const (
	LocalKeeper  = "local"
	RemoteKeeper = "remote"
)

// Keeper represents a generic keeper that allows to read the data of a specific SDK module
type Keeper interface {

	// Type returns whether the keeper is a LocalKeeper or a RemoteKeeper
	Type() string
}
