package config

import (
	"path/filepath"
	"syscall"

	"github.com/lbryio/lbcd/claimtrie/param"
	btcutil "github.com/lbryio/lbcutil"
)

func init() {
	// ensure that we have enough file handles
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		panic("Error getting File Handle RLimit: " + err.Error())
	}
	minRequiredFileHandles := uint64(24000) // seven databases and they each require ~2000, plus a few more to be safe
	if rLimit.Max < minRequiredFileHandles {
		panic("Error increasing File Handle RLimit: increasing file handle limit requires " +
			"unavailable privileges. Allow at least 24000 handles.")
	}
	if rLimit.Cur < minRequiredFileHandles {
		rLimit.Cur = minRequiredFileHandles
		err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
		if err != nil {
			panic("Error setting File Handle RLimit to 24000: " + err.Error())
		}
	}
}

var DefaultConfig = Config{
	Params: param.MainNet,

	RamTrie: true, // as it stands the other trie uses more RAM, more time, and 40GB+ of disk space

	DataDir: filepath.Join(btcutil.AppDataDir("chain", false), "data"),

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
}

// Config is the container of all configurations.
type Config struct {
	Params param.ClaimTrieParams

	RamTrie bool

	DataDir string

	BlockRepoPebble      pebbleConfig
	NodeRepoPebble       pebbleConfig
	TemporalRepoPebble   pebbleConfig
	MerkleTrieRepoPebble pebbleConfig
}

type pebbleConfig struct {
	Path string
}
