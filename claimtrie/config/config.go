package config

import (
	"path/filepath"

	"github.com/btcsuite/btcutil"
)

var (
	// Config is the container for all configurables.
	Config = defaultConfig

	appPath = btcutil.AppDataDir("chain", false)

	defaultConfig = config{
		BlockRepoPebble: pebbleConfig{
			Path: filepath.Join(appPath, "data", "blocks_pebble.db"),
		},
		NodeRepoPebble: pebbleConfig{
			Path: filepath.Join(appPath, "data", "node_change_pebble.db"),
		},
		NodeRepoPostgres: postgresConfig{
			DSN:  "host=localhost user=postgres password=postgres dbname=claimtrie port=5432 sslmode=disable",
			Drop: false,
		},
		TemporalRepoPebble: pebbleConfig{
			Path: filepath.Join(appPath, "data", "temporal_pebble.db"),
		},
		TemporalRepoPostgres: postgresConfig{
			DSN:  "host=localhost user=postgres password=postgres dbname=claimtrie port=5432 sslmode=disable",
			Drop: false,
		},
		MerkleTrieRepoPebble: pebbleConfig{
			Path: filepath.Join(appPath, "data", "merkletrie_pebble.db"),
		},

		ChainRepoPebble: pebbleConfig{
			Path: filepath.Join(appPath, "data", "chain_change_pebble.db"),
		},
		ChainRepoPostgres: postgresConfig{
			DSN:  "host=localhost user=postgres password=postgres dbname=claimtrie port=5432 sslmode=disable",
			Drop: false,
		},
		ReportedBlockRepoPebble: pebbleConfig{
			Path: filepath.Join(appPath, "data", "reported_blocks_pebble.db"),
		},
		TestPostgresDB: postgresConfig{
			DSN:  "host=localhost user=postgres password=postgres dbname=claimtrie_test port=5432 sslmode=disable",
			Drop: true,
		},
	}
)

// config is the container of all configurations.
type config struct {
	BlockRepoPebble      pebbleConfig
	NodeRepoPostgres     postgresConfig
	NodeRepoPebble       pebbleConfig
	TemporalRepoPebble   pebbleConfig
	TemporalRepoPostgres postgresConfig
	MerkleTrieRepoPebble pebbleConfig

	ChainRepoPebble         pebbleConfig
	ChainRepoPostgres       postgresConfig
	ReportedBlockRepoPebble pebbleConfig
	TestPostgresDB          postgresConfig
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
