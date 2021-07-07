package merkletrie

import (
	"io"
)

// Repo defines APIs for MerkleTrie to access persistence layer.
type Repo interface {
	Get(key []byte) ([]byte, io.Closer, error)
	Set(key, value []byte) error
	Close() error
}
