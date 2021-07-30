package noderepo

import (
	"bytes"
	"sort"

	"github.com/btcsuite/btcd/claimtrie/change"
	"github.com/cockroachdb/pebble"
	"github.com/pkg/errors"
)

type Pebble struct {
	db *pebble.DB
}

func NewPebble(path string) (*Pebble, error) {

	db, err := pebble.Open(path, &pebble.Options{Cache: pebble.NewCache(32 << 20), BytesPerSync: 4 << 20})
	repo := &Pebble{db: db}

	return repo, errors.Wrapf(err, "unable to open %s", path)
}

// AppendChanges makes an assumption that anything you pass to it is newer than what was saved before.
func (repo *Pebble) AppendChanges(changes []change.Change) error {

	batch := repo.db.NewBatch()
	defer batch.Close()

	buffer := bytes.NewBuffer(nil)

	for _, chg := range changes {
		buffer.Reset()
		err := chg.MarshalTo(buffer)
		if err != nil {
			return errors.Wrap(err, "in marshaller")
		}

		err = batch.Merge(chg.Name, buffer.Bytes(), pebble.NoSync)
		if err != nil {
			return errors.Wrap(err, "in merge")
		}
	}
	return errors.Wrap(batch.Commit(pebble.NoSync), "in commit")
}

func (repo *Pebble) LoadChanges(name []byte) ([]change.Change, error) {

	data, closer, err := repo.db.Get(name)
	if err != nil && err != pebble.ErrNotFound {
		return nil, errors.Wrapf(err, "in get %s", name) // does returning a name in an error expose too much?
	}
	if closer != nil {
		defer closer.Close()
	}

	return unmarshalChanges(name, data)
}

func unmarshalChanges(name, data []byte) ([]change.Change, error) {
	var changes []change.Change

	buffer := bytes.NewBuffer(data)
	for buffer.Len() > 0 {
		var chg change.Change
		err := chg.UnmarshalFrom(buffer)
		if err != nil {
			return nil, errors.Wrap(err, "in decode")
		}
		chg.Name = name
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
	if err != nil {
		return errors.Wrapf(err, "in load changes for %s", name)
	}
	i := 0
	for ; i < len(changes); i++ {
		if changes[i].Height > finalHeight {
			break
		}
	}
	// making a performance assumption that DropChanges won't happen often:
	err = repo.db.Set(name, []byte{}, pebble.NoSync)
	if err != nil {
		return errors.Wrapf(err, "in set at %s", name)
	}
	return repo.AppendChanges(changes[:i])
}

func (repo *Pebble) IterateChildren(name []byte, f func(changes []change.Change) bool) error {
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
		// NOTE! iter.Key() is ephemeral!
		changes, err := unmarshalChanges(iter.Key(), iter.Value())
		if err != nil {
			return errors.Wrapf(err, "from unmarshaller at %s", iter.Key())
		}
		if !f(changes) {
			break
		}
	}
	return nil
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
		// if we fail to close are we going to try again later?
		return errors.Wrap(err, "on flush")
	}

	err = repo.db.Close()
	return errors.Wrap(err, "on close")
}

func (repo *Pebble) Flush() error {
	_, err := repo.db.AsyncFlush()
	return err
}
