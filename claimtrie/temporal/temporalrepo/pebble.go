package temporalrepo

import (
	"bytes"
	"encoding/binary"
	"fmt"

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

func (repo *Pebble) SetNodeAt(name []byte, height int32) error {

	// key format: height(4B) + 0(1B) + name(varable length)
	// value: nil

	key := bytes.NewBuffer(nil)
	binary.Write(key, binary.BigEndian, height)
	binary.Write(key, binary.BigEndian, byte(0))
	key.Write(name)

	err := repo.db.Set(key.Bytes(), nil, pebble.NoSync)
	if err != nil {
		return fmt.Errorf("pebble set: %w", err)
	}

	return nil
}

func (repo *Pebble) NodesAt(height int32) ([][]byte, error) {

	prefix := bytes.NewBuffer(nil)
	binary.Write(prefix, binary.BigEndian, height)
	binary.Write(prefix, binary.BigEndian, byte(0))

	end := bytes.NewBuffer(nil)
	binary.Write(end, binary.BigEndian, height)
	binary.Write(end, binary.BigEndian, byte(1))

	prefixIterOptions := &pebble.IterOptions{
		LowerBound: prefix.Bytes(),
		UpperBound: end.Bytes(),
	}

	var names [][]byte

	iter := repo.db.NewIter(prefixIterOptions)
	for iter.First(); iter.Valid(); iter.Next() {
		// Skipping the first 5 bytes (height and a null byte), we get the name.
		name := make([]byte, len(iter.Key())-5)
		copy(name, iter.Key()[5:])
		names = append(names, name)
	}

	err := iter.Close()
	if err != nil {
		return nil, fmt.Errorf("pebble get: %w", err)
	}

	return names, nil
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
