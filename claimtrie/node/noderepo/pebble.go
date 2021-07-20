package noderepo

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/claimtrie/change"
	"github.com/btcsuite/btcd/claimtrie/node"
	"github.com/btcsuite/btcd/wire"

	"github.com/cockroachdb/pebble"
	"github.com/vmihailenco/msgpack/v5"
)

type Pebble struct {
	db *pebble.DB
}

func init() {
	claimEncoder := func(e *msgpack.Encoder, v reflect.Value) error {
		claim := v.Interface().(change.ClaimID)
		return e.EncodeBytes(claim[:])
	}
	claimDecoder := func(e *msgpack.Decoder, v reflect.Value) error {
		data, err := e.DecodeBytes()
		if err != nil {
			return err
		}
		if len(data) > change.ClaimIDSize {
			id, err := change.NewIDFromString(string(data))
			if err != nil {
				return err
			}
			v.Set(reflect.ValueOf(id))
		} else {
			id := change.ClaimID{}
			copy(id[:], data)
			v.Set(reflect.ValueOf(id))
		}
		return nil
	}
	msgpack.Register(change.ClaimID{}, claimEncoder, claimDecoder)

	opEncoder := func(e *msgpack.Encoder, v reflect.Value) error {
		op := v.Interface().(wire.OutPoint)
		if err := e.EncodeBytes(op.Hash[:]); err != nil {
			return err
		}
		return e.EncodeUint32(op.Index)
	}
	opDecoder := func(e *msgpack.Decoder, v reflect.Value) error {
		data, err := e.DecodeBytes()
		if err != nil {
			return err
		}
		if len(data) > chainhash.HashSize {
			// try the older data:
			op := node.NewOutPointFromString(string(data))
			v.Set(reflect.ValueOf(*op))
		} else {
			index, err := e.DecodeUint32()
			if err != nil {
				return err
			}
			hash, err := chainhash.NewHash(data)
			if err != nil {
				return err
			}
			op := wire.OutPoint{Hash: *hash, Index: index}
			v.Set(reflect.ValueOf(op))
		}
		return nil
	}
	msgpack.Register(wire.OutPoint{}, opEncoder, opDecoder)
}

func NewPebble(path string) (*Pebble, error) {

	db, err := pebble.Open(path, &pebble.Options{Cache: pebble.NewCache(256 << 20), BytesPerSync: 16 << 20})
	if err != nil {
		return nil, fmt.Errorf("pebble open %s, %w", path, err)
	}

	repo := &Pebble{db: db}

	return repo, nil
}

// AppendChanges makes an assumption that anything you pass to it is newer than what was saved before.
func (repo *Pebble) AppendChanges(changes []change.Change) error {

	batch := repo.db.NewBatch()

	// TODO: switch to buffer pool and reuse encoder
	for _, chg := range changes {
		value, err := msgpack.Marshal(chg)
		if err != nil {
			return fmt.Errorf("msgpack marshal value: %w", err)
		}

		err = batch.Merge(chg.Name, value, pebble.NoSync)
		if err != nil {
			return fmt.Errorf("pebble set: %w", err)
		}
	}
	err := batch.Commit(pebble.NoSync)
	if err != nil {
		return fmt.Errorf("pebble save commit: %w", err)
	}
	batch.Close()
	return err
}

func (repo *Pebble) LoadChanges(name []byte) ([]change.Change, error) {

	data, closer, err := repo.db.Get(name)
	if err != nil && err != pebble.ErrNotFound {
		return nil, fmt.Errorf("pebble get: %w", err)
	}
	if closer != nil {
		defer closer.Close()
	}

	return unmarshalChanges(data)
}

func unmarshalChanges(data []byte) ([]change.Change, error) {
	var changes []change.Change
	dec := msgpack.GetDecoder()
	defer msgpack.PutDecoder(dec)

	reader := bytes.NewReader(data)
	dec.Reset(reader)
	for reader.Len() > 0 {
		var chg change.Change
		err := dec.Decode(&chg)
		if err != nil {
			return nil, fmt.Errorf("msgpack unmarshal: %w", err)
		}
		changes = append(changes, chg)
	}

	// this was required for the normalization stuff:
	sort.SliceStable(changes, func(i, j int) bool {
		return changes[i].Height < changes[j].Height
	})

	return changes, nil
}

func (repo *Pebble) DropChanges(name []byte, finalHeight int32) error {
	changes, err := repo.LoadChanges(name)
	i := 0
	for ; i < len(changes); i++ {
		if changes[i].Height > finalHeight {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("pebble drop: %w", err)
	}
	// making a performance assumption that DropChanges won't happen often:
	err = repo.db.Set(name, []byte{}, pebble.NoSync)
	if err != nil {
		return fmt.Errorf("pebble drop: %w", err)
	}
	return repo.AppendChanges(changes[:i])
}

func (repo *Pebble) IterateChildren(name []byte, f func(changes []change.Change) bool) {
	start := make([]byte, len(name)+1) // zeros that last byte; need a constant len for stack alloc?
	copy(start, name)

	end := make([]byte, 256) // max name length is 255
	copy(end, name)
	for i := len(name); i < 256; i++ {
		end[i] = 255
	}

	prefixIterOptions := &pebble.IterOptions{
		LowerBound: start,
		UpperBound: end,
	}

	iter := repo.db.NewIter(prefixIterOptions)
	defer iter.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		changes, err := unmarshalChanges(iter.Value())
		if err != nil {
			panic(err)
		}
		if !f(changes) {
			return
		}
	}
}

func (repo *Pebble) IterateAll(predicate func(name []byte) bool) {
	iter := repo.db.NewIter(nil)
	defer iter.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		if !predicate(iter.Key()) {
			break
		}
	}
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
