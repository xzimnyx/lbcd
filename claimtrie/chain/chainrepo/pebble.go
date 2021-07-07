package chainrepo

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/btcsuite/btcd/claimtrie/change"
	"github.com/vmihailenco/msgpack/v5"

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

func (repo *Pebble) Save(height int32, changes []change.Change) error {

	if len(changes) == 0 {
		return nil
	}

	key := bytes.NewBuffer(nil)
	err := binary.Write(key, binary.BigEndian, height)
	if err != nil {
		return fmt.Errorf("pebble prepare key: %w", err)
	}

	value, err := msgpack.Marshal(changes)
	if err != nil {
		return fmt.Errorf("pebble msgpack marshal: %w", err)
	}

	err = repo.db.Set(key.Bytes(), value, pebble.NoSync)
	if err != nil {
		return fmt.Errorf("pebble set: %w", err)
	}

	return nil
}

func (repo *Pebble) Load(height int32) ([]change.Change, error) {

	key := bytes.NewBuffer(nil)
	err := binary.Write(key, binary.BigEndian, height)
	if err != nil {
		return nil, fmt.Errorf("pebble prepare key: %w", err)
	}

	b, closer, err := repo.db.Get(key.Bytes())
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	var changes []change.Change
	err = msgpack.Unmarshal(b, &changes)
	if err != nil {
		return nil, fmt.Errorf("pebble msgpack marshal: %w", err)
	}

	return changes, nil
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
