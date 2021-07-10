package cmd

import (
	"fmt"
	"log"
	"strconv"

	"github.com/btcsuite/btcd/claimtrie/block/blockrepo"
	"github.com/btcsuite/btcd/claimtrie/merkletrie"
	"github.com/btcsuite/btcd/claimtrie/merkletrie/merkletrierepo"
	"github.com/btcsuite/btcd/claimtrie/param"
	"github.com/btcsuite/btcd/claimtrie/temporal/temporalrepo"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(blockCmd)

	blockCmd.AddCommand(blockLastCmd)
	blockCmd.AddCommand(blockListCmd)
	blockCmd.AddCommand(blockNameCmd)
}

var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "Block related commands",
}

var blockLastCmd = &cobra.Command{
	Use:   "last",
	Short: "Show the Merkle Hash of the last block",
	RunE: func(cmd *cobra.Command, args []string) error {

		repo, err := blockrepo.NewPebble(localConfig.ReportedBlockRepoPebble.Path)
		if err != nil {
			log.Fatalf("can't open reported block repo: %s", err)
		}

		last, err := repo.Load()
		if err != nil {
			return fmt.Errorf("load previous height")
		}

		hash, err := repo.Get(last)
		if err != nil {
			return fmt.Errorf("load changes from repo: %w", err)
		}

		fmt.Printf("blk %-7d: %s\n", last, hash.String())

		return nil
	},
}

var blockListCmd = &cobra.Command{
	Use:   "list <from_height> [<to_height>]",
	Short: "List the Merkle Hash of block in a range of heights",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {

		repo, err := blockrepo.NewPebble(localConfig.ReportedBlockRepoPebble.Path)
		if err != nil {
			log.Fatalf("can't open reported block repo: %s", err)
		}

		fromHeight, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid args")
		}

		toHeight := fromHeight + 1
		if len(args) == 2 {
			toHeight, err = strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid args")
			}
		}

		last, err := repo.Load()
		if err != nil {
			return fmt.Errorf("load previous height")
		}

		if toHeight >= int(last) {
			toHeight = int(last)
		}

		for i := fromHeight; i < toHeight; i++ {
			hash, err := repo.Get(int32(i))
			if err != nil {
				return fmt.Errorf("load changes from repo: %w", err)
			}
			fmt.Printf("blk %-7d: %s\n", i, hash.String())
		}

		return nil
	},
}

var blockNameCmd = &cobra.Command{
	Use:   "vertex <height> <name>",
	Short: "List the claim and child hashes at vertex name of block at height",
	Args:  cobra.RangeArgs(2, 2),
	RunE: func(cmd *cobra.Command, args []string) error {

		repo, err := blockrepo.NewPebble(localConfig.BlockRepoPebble.Path)
		if err != nil {
			return fmt.Errorf("can't open reported block repo: %w", err)
		}
		defer repo.Close()

		height, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid args")
		}

		last, err := repo.Load()
		if err != nil {
			return fmt.Errorf("load previous height: %w", err)
		}

		if last < int32(height) {
			return fmt.Errorf("requested height is unavailable")
		}

		hash, err := repo.Get(int32(height))
		if err != nil {
			return fmt.Errorf("load previous height: %w", err)
		}

		trieRepo, err := merkletrierepo.NewPebble(localConfig.MerkleTrieRepoPebble.Path)
		if err != nil {
			return fmt.Errorf("can't open merkle trie repo: %w", err)
		}

		trie := merkletrie.New(nil, trieRepo)
		defer trie.Close()
		trie.SetRoot(hash)
		if len(args) > 1 {
			trie.Dump(args[1], param.AllClaimsInMerkleForkHeight >= int32(height))
		} else {
			tmpRepo, err := temporalrepo.NewPebble(localConfig.TemporalRepoPebble.Path)
			if err != nil {
				return fmt.Errorf("can't open temporal repo: %w", err)
			}
			nodes, err := tmpRepo.NodesAt(int32(height))
			if err != nil {
				return fmt.Errorf("can't read temporal repo at %d: %w", height, err)
			}
			for _, name := range nodes {
				fmt.Printf("Name: %s, ", string(name))
				trie.Dump(string(name), param.AllClaimsInMerkleForkHeight >= int32(height))
			}
		}
		return nil
	},
}
