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

package config

import (
	"path/filepath"

	"github.com/btcsuite/btcutil"
)

var (
	// Config is the container for all configurables.
	//
	// TODO: enable overriding using config.yaml or environment variables.
	Config = defaultConfig

	lbrycrdPath = btcutil.AppDataDir("chain", false)

	defaultConfig = config{
		BlockRepo: pebbleConfig{
			Path: filepath.Join(lbrycrdPath, "data", "blocks.db"),
		},
		ChangeRepo: postgresConfig{
			DSN:  "host=localhost user=postgres password=postgres dbname=claimtrie port=5432 sslmode=disable",
			Drop: true,
		},
		NodeRepo: postgresConfig{
			DSN:  "host=localhost user=postgres password=postgres dbname=claimtrie port=5432 sslmode=disable",
			Drop: true,
		},
		TemporalRepo: pebbleConfig{
			Path: filepath.Join(lbrycrdPath, "data", "temporal.db"),
		},
		TrieRepo: pebbleConfig{
			Path: filepath.Join(lbrycrdPath, "data", "merkletrie.db"),
		},
		ReportedBlockRepo: pebbleConfig{
			Path: filepath.Join(lbrycrdPath, "data", "reported_blocks.db"),
		},
		TestPostgresDB: postgresConfig{
			DSN:  "host=localhost user=postgres password=postgres dbname=claimtrie_test port=5432 sslmode=disable",
			Drop: true,
		},
	}
)

// config is the container of all configurations.
type config struct {
	BlockRepo    pebbleConfig
	ChangeRepo   postgresConfig
	NodeRepo     postgresConfig
	TemporalRepo pebbleConfig
	TrieRepo     pebbleConfig

	ReportedBlockRepo pebbleConfig
	TestPostgresDB    postgresConfig
}

type postgresConfig struct {
	// Data source name
	DSN string
	// Drop tables at loading. Set true for debugging.
	Drop bool
}

type pebbleConfig struct {
	Path string
}
