package chainrepo

import (
	"fmt"

	"github.com/btcsuite/btcd/claimtrie/change"

	"github.com/cockroachdb/pebble"
)

type Pebble struct {
	db *pebble.DB
}

func NewPebble(path string) (*Pebble, error) {

	db, err := pebble.Open(path, nil)
	if err != nil {
		return nil, fmt.Errorf("pebble open %s, %w", path, err)
	}

	repo := &Pebble{db: db}

	return repo, nil
}

func (repo *Pebble) Save(changes []change.Change) error {

	// TODO

	return nil
}

func (repo *Pebble) LoadByHeight(height int32) ([]change.Change, error) {

	// TODO: should change the to stream-like API, such as reader, iterator, etc.

	return nil, nil
}

func (repo *Pebble) Close() error {

	err := repo.db.Flush()
	if err != nil {
		return fmt.Errorf("pebble fludh: %w", err)
	}

	err = repo.db.Close()
	if err != nil {
		return fmt.Errorf("pebble close: %w", err)
	}

	return nil
}
