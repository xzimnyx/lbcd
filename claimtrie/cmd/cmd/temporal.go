package cmd

import (
	"fmt"
	"log"
	"strconv"

	"github.com/btcsuite/btcd/claimtrie/temporal/temporalrepo"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(temporalCmd)
}

var temporalCmd = &cobra.Command{
	Use:   "temporal <from_height> [<to_height>]]",
	Short: "List which nodes are update in a range of heights",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runListNodes,
}

func runListNodes(cmd *cobra.Command, args []string) error {

	repo, err := temporalrepo.NewPebble(localConfig.TemporalRepoPebble.Path)
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

	for height := fromHeight; height < toHeight; height++ {
		names, err := repo.NodesAt(int32(height))
		if err != nil {
			return fmt.Errorf("get node names from temporal")
		}

		if len(names) == 0 {
			continue
		}

		fmt.Printf("%7d: %q", height, names[0])
		for _, name := range names[1:] {
			fmt.Printf(", %q ", name)
		}
		fmt.Printf("\n")
	}

	return nil
}
