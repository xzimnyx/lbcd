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

package node

import "github.com/btcsuite/btcd/claimtrie/merkletrie"

type NodeManager struct {
	repo NodeRepo

	cache map[string]*Node
}

func NewNodeManager(repo NodeRepo) (*NodeManager, error) {

	nm := &NodeManager{
		repo:  repo,
		cache: map[string]*Node{},
	}

	return nm, nil
}

func (nm *NodeManager) Get(key []byte) merkletrie.Value {
	return nm.GetNode(string(key))
}

func (nm *NodeManager) GetNode(name string) *Node {
	n, ok := nm.cache[name]
	if !ok {
		n = NewNode()
		nm.cache[name] = n
	}
	return n
}

func (nm *NodeManager) Close() error {
	return nm.repo.Close()
}
