// Copyright (c) 2021 - LBRY Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

	key, _ := msgpack.Marshal(height)
	value, _ := msgpack.Marshal(names)

	return repo.db.Set(key, value, pebble.NoSync)
}

func (repo *TemporalRepoPebble) Close() error {
	return repo.db.Close()
}
