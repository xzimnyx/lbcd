package node

import "github.com/btcsuite/btcd/claimtrie/change"

// Repo defines APIs for Node to access persistence layer.
type Repo interface {
	// SaveChanges saves changes into the repo.
	// The changes can belong to different nodes, but the chronological
	// order must be preserved for the same node.
	SaveChanges(changes []change.Change) error

	// LoadChanges loads changes of a node up to (includes) the specified height.
	// If no changes found, both returned slice and error will be nil.
	LoadChanges(name []byte, height int32) ([]change.Change, error)

	// Close closes the repo.
	Close() error
}
