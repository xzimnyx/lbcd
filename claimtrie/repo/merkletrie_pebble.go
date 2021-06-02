package repo

import (
	"fmt"
	"io"

	"github.com/cockroachdb/pebble"
)

type MerkleTrieRepoPebble struct {
	db *pebble.DB
}

func NewMerkleTrieRepoPebble(path string) (*MerkleTrieRepoPebble, error) {

	db, err := pebble.Open(path, nil)
	if err != nil {
		return nil, fmt.Errorf("pebble open %s, %w", path, err)
	}

	repo := &MerkleTrieRepoPebble{
		db: db,
	}

	return repo, nil
}

func (repo *MerkleTrieRepoPebble) Get(key []byte) ([]byte, io.Closer, error) {
	return repo.db.Get(key)
}

func (repo *MerkleTrieRepoPebble) Set(key, value []byte) error {
	return repo.db.Set(key, value, pebble.NoSync)
}

func (repo *MerkleTrieRepoPebble) Close() error {

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
