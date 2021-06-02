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
		ChainChangeRepoPebble: pebbleConfig{
			Path: filepath.Join(appPath, "data", "chain_change_pebble.db"),
		},
		ChainChangeRepoPostgres: postgresConfig{
			DSN:  "host=localhost user=postgres password=postgres dbname=claimtrie port=5432 sslmode=disable",
			Drop: false,
		},
		NodeChangeRepoPebble: pebbleConfig{
			Path: filepath.Join(appPath, "data", "node_change_pebble.db"),
		},
		NodeChangeRepoPostgres: postgresConfig{
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
		TrieRepoPebble: pebbleConfig{
			Path: filepath.Join(appPath, "data", "merkletrie_pebble.db"),
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
	BlockRepoPebble         pebbleConfig
	ChainChangeRepoPebble   pebbleConfig
	ChainChangeRepoPostgres postgresConfig

	NodeChangeRepoPostgres postgresConfig
	NodeChangeRepoPebble   pebbleConfig
	TemporalRepoPebble     pebbleConfig
	TemporalRepoPostgres   postgresConfig
	TrieRepoPebble         pebbleConfig

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
