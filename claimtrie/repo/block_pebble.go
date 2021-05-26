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
	"encoding/binary"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/pebble"
)

type BlockRepoPebble struct {
	db *pebble.DB
}

func NewBlockRepoPebble(path string) (*BlockRepoPebble, error) {

	db, err := pebble.Open(path, nil)
	if err != nil {
		return nil, fmt.Errorf("pebble open %s, %w", path, err)
	}

	repo := &BlockRepoPebble{db: db}

	return repo, nil
}

func (repo *BlockRepoPebble) Get(height int32) (*chainhash.Hash, error) {

	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, uint32(height))

	b, closer, err := repo.db.Get(key)
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	hash, err := chainhash.NewHash(b)

	return hash, err
}

func (repo *BlockRepoPebble) Set(height int32, hash *chainhash.Hash) error {

	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, uint32(height))

	return repo.db.Set(key, hash[:], pebble.NoSync)
}

func (repo *BlockRepoPebble) Close() error {
	return repo.db.Close()
}
