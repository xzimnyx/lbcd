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
	"testing"

	"github.com/btcsuite/btcd/claimtrie/temporal"
	"github.com/stretchr/testify/assert"
)

func TestTemporalRepoMem(t *testing.T) {

	repo := NewTemporalMem()
	testTemporalRepo(t, repo)
}

func TestTemporalRepoPebble(t *testing.T) {

	repo, err := NewTemporalPebble(cfg.TemporalRepo.Path)
	if assert.NoError(t, err) {
		testTemporalRepo(t, repo)
	}
}

func testTemporalRepo(t *testing.T, repo temporal.TemporalRepo) {

	testcases := []struct {
		name    string
		heights []int32
	}{
		{"a", []int32{1, 3, 2}},
		{"a", []int32{2, 3}},
		{"b", []int32{5, 4}},
		{"b", []int32{5, 1}},
		{"c", []int32{4, 3, 8}},
	}

	for _, i := range testcases {
		for _, height := range i.heights {
			err := repo.SetNodeAt(i.name, height)
			assert.NoError(t, err)
		}
	}

	// a: 1, 2, 3
	// b: 1, 5, 4
	// c: 4, 3, 8

	names, err := repo.NodesAt(2)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"a"}, names)

	names, err = repo.NodesAt(5)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"b"}, names)

	names, err = repo.NodesAt(8)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"c"}, names)

	names, err = repo.NodesAt(1)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"a", "b"}, names)

	names, err = repo.NodesAt(4)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"b", "c"}, names)

	names, err = repo.NodesAt(3)
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "c"}, names)
}
