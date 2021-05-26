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

import "github.com/btcsuite/btcd/wire"

type list map[wire.OutPoint]*Claim

type comparator func(c *Claim) bool

func byID(id string) comparator {
	return func(c *Claim) bool {
		return c.ClaimID == id
	}
}

func byStatus(st Status) comparator {
	return func(c *Claim) bool {
		return c.Status == st
	}
}

func (l list) removeAll(cmp comparator) {

	for op, v := range l {
		if cmp(v) {
			delete(l, op)
		}
	}
}

func (l list) find(cmp comparator) *Claim {

	for _, v := range l {
		if cmp(v) {
			return v
		}
	}

	return nil
}
