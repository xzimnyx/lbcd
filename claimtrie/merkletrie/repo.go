package merkletrie

import (
	"io"
)

type Repo interface {
	Get(key []byte) ([]byte, io.Closer, error)
	Set(key, value []byte) error
	Close() error
}
