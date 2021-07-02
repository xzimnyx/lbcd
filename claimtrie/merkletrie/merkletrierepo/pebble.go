package merkletrierepo

import (
	"fmt"
	"io"
	"time"

	"github.com/cockroachdb/pebble"
	humanize "github.com/dustin/go-humanize"
)

type Pebble struct {
	db *pebble.DB
}

func NewPebble(path string) (*Pebble, error) {

	cache := pebble.NewCache(512 << 20)
	defer cache.Unref()

	go func() {
		tick := time.NewTicker(60 * time.Second)
		for range tick.C {

			m := cache.Metrics()
			fmt.Printf("cnt: %s, objs: %s, hits: %s, miss: %s, hitrate: %.2f\n",
				humanize.Bytes(uint64(m.Size)),
				humanize.Comma(m.Count),
				humanize.Comma(m.Hits),
				humanize.Comma(m.Misses),
				float64(m.Hits)/float64(m.Hits+m.Misses))

		}
	}()

	db, err := pebble.Open(path, &pebble.Options{Cache: cache})
	if err != nil {
		return nil, fmt.Errorf("pebble open %s, %w", path, err)
	}

	repo := &Pebble{
		db: db,
	}

	return repo, nil
}

func (repo *Pebble) Get(key []byte) ([]byte, io.Closer, error) {
	return repo.db.Get(key)
}

func (repo *Pebble) Set(key, value []byte) error {
	return repo.db.Set(key, value, pebble.NoSync)
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
