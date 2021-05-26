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

package param

const (
	DefaultMaxActiveDelay    int32 = 4032
	DefaultActiveDelayFactor int32 = 32
)

// https://lbry.io/news/hf1807
const (
	DefaultOriginalClaimExpirationTime       int32 = 262974
	DefaultExtendedClaimExpirationTime       int32 = 2102400
	DefaultExtendedClaimExpirationForkHeight int32 = 400155
)

var (
	MaxActiveDelay                    = DefaultMaxActiveDelay
	ActiveDelayFactor                 = DefaultActiveDelayFactor
	OriginalClaimExpirationTime       = DefaultOriginalClaimExpirationTime
	ExtendedClaimExpirationTime       = DefaultExtendedClaimExpirationTime
	ExtendedClaimExpirationForkHeight = DefaultExtendedClaimExpirationForkHeight
)

// https://lbry.com/news/hf1903

// https://lbry.com/news/hf1910
