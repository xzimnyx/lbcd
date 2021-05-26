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

type TemporalRepoMem struct {
	cache map[int32]map[string]bool
}

func NewTemporalMem() *TemporalRepoMem {
	return &TemporalRepoMem{
		cache: map[int32]map[string]bool{},
	}
}

func (repo *TemporalRepoMem) NodesAt(height int32) ([]string, error) {

	var names []string

	for name := range repo.cache[height] {
		names = append(names, name)
	}

	return names, nil
}

func (repo *TemporalRepoMem) SetNodeAt(name string, height int32) error {

	names, ok := repo.cache[height]
	if !ok {
		names = map[string]bool{}
		repo.cache[height] = names
	}
	names[name] = true

	return nil
}

func (repo *TemporalRepoMem) Close() error {
	return nil
}
