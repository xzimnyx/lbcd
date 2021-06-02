package block

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

type Repo interface {
	Load() (int32, error)
	Set(height int32, hash *chainhash.Hash) error
	Get(height int32) (*chainhash.Hash, error)
	Close() error
}
