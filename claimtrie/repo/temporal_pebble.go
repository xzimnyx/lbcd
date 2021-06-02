package repo

import (
	"fmt"

	"github.com/cockroachdb/pebble"
	"github.com/vmihailenco/msgpack/v5"
)

type TemporalRepoPebble struct {
	db *pebble.DB
}

func NewTemporalPebble(path string) (*TemporalRepoPebble, error) {

	db, err := pebble.Open(path, nil)
	if err != nil {
		return nil, fmt.Errorf("pebble open %s, %w", path, err)
	}

	repo := &TemporalRepoPebble{db: db}

	return repo, nil
}

func (repo *TemporalRepoPebble) NodesAt(height int32) ([]string, error) {

	key, err := msgpack.Marshal(height)
	if err != nil {
		return nil, fmt.Errorf("msgpack marshal: %w", err)
	}

	value, closer, err := repo.db.Get(key)
	if err == pebble.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	var names []string
	err = msgpack.Unmarshal(value, &names)
	if err != nil {
		return nil, fmt.Errorf("msgpack unmarshal: %w", err)
	}

	return names, nil
}

func (repo *TemporalRepoPebble) SetNodeAt(name string, height int32) error {

	names, err := repo.NodesAt(height)
	if err != nil && err != pebble.ErrNotFound {
		return fmt.Errorf("get repo: %w", err)
	}

	for _, n := range names {
		if n == name {
			return nil
		}
	}
	names = append(names, name)

	key, err := msgpack.Marshal(height)
	if err != nil {
		return fmt.Errorf("msgpack marshal key: %w", err)
	}

	value, err := msgpack.Marshal(names)
	if err != nil {
		return fmt.Errorf("msgpack marshal value: %w", err)
	}

	return repo.db.Set(key, value, pebble.NoSync)
}

func (repo *TemporalRepoPebble) Close() error {

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
