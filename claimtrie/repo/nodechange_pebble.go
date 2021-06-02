package repo

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/btcsuite/btcd/claimtrie/change"
	"github.com/cockroachdb/pebble"
	"github.com/vmihailenco/msgpack/v5"
)

type NodeChangeRepoPebble struct {
	db *pebble.DB
}

func NewNodeChangeRepoPebble(path string) (*NodeChangeRepoPebble, error) {

	db, err := pebble.Open(path, nil)
	if err != nil {
		return nil, fmt.Errorf("pebble open %s, %w", path, err)
	}

	repo := &NodeChangeRepoPebble{db: db}

	return repo, nil
}

func (repo *NodeChangeRepoPebble) Save(changes []change.Change) error {

	// Instead of load-modify-save, we always write individual change.
	// To preserve the chronological order of changes, we encode the key in:
	//
	//    name height index in the changes slice.
	//
	for i, chg := range changes {

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

func (repo *NodeChangeRepoPebble) LoadByNameUpToHeight(name string, height int32) ([]change.Change, error) {

	keyUpperBound := func(b []byte) []byte {
		end := make([]byte, len(b))
		copy(end, b)
		for i := len(end) - 1; i >= 0; i-- {
			end[i] = end[i] + 1
			if end[i] != 0 {
				return end[:i+1]
			}
		}
		return nil // no upper-bound
	}

	prefixIterOptions := func(prefix []byte) *pebble.IterOptions {
		return &pebble.IterOptions{
			LowerBound: prefix,
			UpperBound: keyUpperBound(prefix),
		}
	}

	var changes []change.Change

	prefix := []byte(name)
	prefix = append(prefix, 0)

	iter := repo.db.NewIter(prefixIterOptions(prefix))
	for iter.First(); iter.Valid(); iter.Next() {
		var chg change.Change
		err := msgpack.Unmarshal(iter.Value(), &chg)
		if err != nil {
			return nil, fmt.Errorf("msgpack unmarshal value: %w", err)
		}
		if chg.Height <= height {
			changes = append(changes, chg)
		}
	}

	err := iter.Close()
	if err != nil {
		return nil, fmt.Errorf("pebble get: %w", err)
	}

	return changes, nil
}

func (repo *NodeChangeRepoPebble) Close() error {

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
