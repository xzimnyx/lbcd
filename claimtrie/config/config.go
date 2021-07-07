package config

import (
	"path/filepath"
)

func GenerateConfig(folder string) *DBConfig {
	return &DBConfig{
		BlockRepoPebble: pebbleConfig{
			Path: filepath.Join(folder, "blocks_pebble_db"),
		},
		NodeRepoPebble: pebbleConfig{
			Path: filepath.Join(folder, "node_change_pebble_db"),
		},
		TemporalRepoPebble: pebbleConfig{
			Path: filepath.Join(folder, "temporal_pebble_db"),
		},
		MerkleTrieRepoPebble: pebbleConfig{
			Path: filepath.Join(folder, "merkletrie_pebble_db"),
		},
		ChainRepoPebble: pebbleConfig{
			Path: filepath.Join(folder, "chain_pebble_db"),
		},
		ReportedBlockRepoPebble: pebbleConfig{
			Path: filepath.Join(folder, "reported_blocks_pebble_db"),
		},
	}
}

// DBConfig is the container of all configurations.
type DBConfig struct {
	BlockRepoPebble      pebbleConfig
	NodeRepoPebble       pebbleConfig
	TemporalRepoPebble   pebbleConfig
	MerkleTrieRepoPebble pebbleConfig

	ChainRepoPebble         pebbleConfig
	ReportedBlockRepoPebble pebbleConfig
}

type pebbleConfig struct {
	Path string
}
