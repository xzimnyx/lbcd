package noderepo

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/btcsuite/btcd/claimtrie/change"
	"github.com/cockroachdb/pebble"
	"github.com/vmihailenco/msgpack/v5"
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

func (repo *Pebble) SaveChanges(changes []change.Change) error {

	for i, chg := range changes {

		// Key format: name(variable size) + 0(1B) + height(4B) + 0 + slice_idx (4B)
		// The last slice_idx is just to keep the entry unique and preserve the order.
		buf := bytes.NewBuffer(nil)
		buf.Write(chg.Name)
		binary.Write(buf, binary.BigEndian, byte(0))
		binary.Write(buf, binary.BigEndian, chg.Height)
		binary.Write(buf, binary.BigEndian, byte(0))
		binary.Write(buf, binary.BigEndian, int32(i))

		value, err := msgpack.Marshal(chg)
		if err != nil {
			return fmt.Errorf("msgpack marshal value: %w", err)
		}

		err = repo.db.Set(buf.Bytes(), value, pebble.NoSync)
		if err != nil {
			return fmt.Errorf("pebble set: %w", err)
		}
	}

	return nil
}

func (repo *Pebble) LoadChanges(name []byte, height int32) ([]change.Change, error) {

	prefix := bytes.NewBuffer(nil)
	prefix.Write(name)
	binary.Write(prefix, binary.BigEndian, byte(0))

	end := bytes.NewBuffer(nil)
	end.Write(name)
	binary.Write(end, binary.BigEndian, byte(0))
	binary.Write(end, binary.BigEndian, height)
	binary.Write(end, binary.BigEndian, byte(1))

	prefixIterOptions := &pebble.IterOptions{
		LowerBound: prefix.Bytes(),
		UpperBound: end.Bytes(),
	}

	var changes []change.Change

	iter := repo.db.NewIter(prefixIterOptions)
	for iter.First(); iter.Valid(); iter.Next() {

		var chg change.Change

		err := msgpack.Unmarshal(iter.Value(), &chg)
		if err != nil {
			return nil, fmt.Errorf("msgpack unmarshal value: %w", err)
		}

		changes = append(changes, chg)
	}

	err := iter.Close()
	if err != nil {
		return nil, fmt.Errorf("pebble get: %w", err)
	}

	return changes, nil
}

func (repo *Pebble) Close() error {

	err := repo.db.Flush()
	if err != nil {
		return fmt.Errorf("pebble flush: %w", err)
	}

	err = repo.db.Close()
	if err != nil {
		return fmt.Errorf("pebble close: %w", err)
	}

	return nil
}
