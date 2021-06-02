package node

import "github.com/btcsuite/btcd/claimtrie/change"

type Repo interface {
	Save(changes []change.Change) error
	LoadByNameUpToHeight(name string, height int32) ([]change.Change, error)
	Close() error
}
