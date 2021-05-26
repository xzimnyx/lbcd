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

package merkletrie

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

type node struct {
	hash     *chainhash.Hash
	links    [256]*node
	hasValue bool
}

func newNode() *node {
	return &node{}
}

// nbuf decodes the on-disk format of a node, which has the following form:
//   ch(1B) hash(32B)
//   ...
//   ch(1B) hash(32B)
//   vhash(32B)
type nbuf []byte

func (nb nbuf) entries() int {
	return len(nb) / 33
}

func (nb nbuf) entry(i int) (byte, *chainhash.Hash) {
	h := chainhash.Hash{}
	copy(h[:], nb[33*i+1:])
	return nb[33*i], &h
}

func (nb nbuf) hasValue() bool {
	return len(nb)%33 == 32
}
