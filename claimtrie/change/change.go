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

package change

const (
	// Node change
	AddClaim     = "AddClaim"
	SpendClaim   = "SpendClaim"
	UpdateClaim  = "UpdateClaim"
	AddSupport   = "AddSupport"
	SpendSupport = "SpendSupport"
)

type Change struct {
	ID     uint   `gorm:"primarykey;index:,type:brin"`
	Type   string `gorm:"index"`
	Height int32  `gorm:"index:,type:brin"`

	Name     []byte `gorm:"index,type:hash"`
	ClaimID  string
	OutPoint string
	Amount   int64
	Value    []byte
}
