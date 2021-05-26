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

	blockCmd.AddCommand(blockLastCmd)
	blockCmd.AddCommand(blockListCmd)
}

var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "Block related commands",
}

var blockLastCmd = &cobra.Command{
	Use:   "last",
	Short: "Show the Merkle Hashlast of the last block",
	RunE: func(cmd *cobra.Command, args []string) error {

		repo, err := blockrepo.NewPebble(config.Config.ReportedBlockRepoPebble.Path)
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

		repo, err := blockrepo.NewPebble(config.Config.ReportedBlockRepoPebble.Path)
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
