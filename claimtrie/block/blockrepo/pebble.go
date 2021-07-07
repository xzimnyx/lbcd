package blockrepo

import (
	"encoding/binary"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"

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

func (repo *Pebble) Load() (int32, error) {

	iter := repo.db.NewIter(nil)
	if !iter.Last() {
		if err := iter.Close(); err != nil {
			return 0, fmt.Errorf("close iter: %w", err)
		}
		return 0, nil
	}

	height := int32(binary.BigEndian.Uint32(iter.Key()))
	if err := iter.Close(); err != nil {
		return height, fmt.Errorf("close iter: %w", err)
	}

	return height, nil
}

func (repo *Pebble) Get(height int32) (*chainhash.Hash, error) {

	key := make([]byte, 4)
	binary.BigEndian.PutUint32(key, uint32(height))

	b, closer, err := repo.db.Get(key)
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	hash, err := chainhash.NewHash(b)

	return hash, err
}

func (repo *Pebble) Set(height int32, hash *chainhash.Hash) error {

	key := make([]byte, 4)
	binary.BigEndian.PutUint32(key, uint32(height))

	return repo.db.Set(key, hash[:], pebble.NoSync)
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
