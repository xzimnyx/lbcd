package chain

import "github.com/btcsuite/btcd/claimtrie/change"

// TODO: swtich to steam-like read API, such as iterartor, or scanner.
type Repo interface {
	Save(changes []change.Change) error
	LoadByHeight(height int32) ([]change.Change, error)
	Close() error
}
