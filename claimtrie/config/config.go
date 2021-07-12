package config

import (
	"path/filepath"

	"github.com/btcsuite/btcutil"
)

var DefaultConfig = Config{
	Record:  false,
	RamTrie: false,

	DataDir: filepath.Join(btcutil.AppDataDir("chain", false), "data", "mainnet", "claim_dbs"),

	BlockRepoPebble: pebbleConfig{
		Path: "blocks_pebble_db",
	},
	NodeRepoPebble: pebbleConfig{
		Path: "node_change_pebble_db",
	},
	TemporalRepoPebble: pebbleConfig{
		Path: "temporal_pebble_db",
	},
	MerkleTrieRepoPebble: pebbleConfig{
		Path: "merkletrie_pebble_db",
	},
	ChainRepoPebble: pebbleConfig{
		Path: "chain_pebble_db",
	},
	ReportedBlockRepoPebble: pebbleConfig{
		Path: "reported_blocks_pebble_db",
	},
}

// Config is the container of all configurations.
type Config struct {
	Record  bool
	RamTrie bool

	DataDir string

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
