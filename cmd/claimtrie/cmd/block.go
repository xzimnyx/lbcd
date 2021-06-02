package cmd

import (
	"fmt"
	"log"
	"strconv"

	"github.com/btcsuite/btcd/claimtrie/block/blockrepo"
	"github.com/btcsuite/btcd/claimtrie/config"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(blockCmd)
}

var blockCmd = &cobra.Command{
	Use:   "block <from_height> [<to_height>]",
	Short: "Show merkle hash of block at height",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  listHash,
}

func listHash(cmd *cobra.Command, args []string) error {

	blockRepo, err := blockrepo.NewPebble(config.Config.ReportedBlockRepoPebble.Path)
	if err != nil {
		log.Fatalf("can't open reported block repo: %s", err)
	}

	blockRepo.Load()

	from, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid args")
	}

	to := from
	if len(args) == 2 {
		to, err = strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid args")
		}
	}

	for i := from; i <= to; i++ {
		hash, err := blockRepo.Get(int32(i))
		if err != nil {
			return fmt.Errorf("load changes from repo: %w", err)
		}
		fmt.Printf("blk %-7d: %s\n", i, hash.String())
	}

	return nil
}
