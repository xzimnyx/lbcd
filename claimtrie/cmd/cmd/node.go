package cmd

import (
	"fmt"
	"math"
	"strconv"

	"github.com/btcsuite/btcd/claimtrie/node"
	"github.com/btcsuite/btcd/claimtrie/node/noderepo"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(nodeCmd)

	nodeCmd.AddCommand(nodeDumpCmd)
	nodeCmd.AddCommand(nodeReplayCmd)
}

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Replay the application of changes on a node up to certain height",
}

var nodeDumpCmd = &cobra.Command{
	Use:   "dump <node_name> [<height>]",
	Short: "Replay the application of changes on a node up to certain height",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {

		repo, err := noderepo.NewPebble(localConfig.NodeRepoPebble.Path)
		if err != nil {
			return fmt.Errorf("open node repo: %w", err)
		}

		name := args[0]
		height := math.MaxInt32

		if len(args) == 2 {
			height, err = strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid args")
			}
		}

		changes, err := repo.LoadChanges([]byte(name))
		if err != nil {
			return fmt.Errorf("load commands: %w", err)
		}

		for _, chg := range changes {
			if int(chg.Height) > height {
				break
			}
			showChange(chg)
		}

		return nil
	},
}

var nodeReplayCmd = &cobra.Command{
	Use:   "replay <node_name> [<height>]",
	Short: "Replay the application of changes on a node up to certain height",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {

		repo, err := noderepo.NewPebble(localConfig.NodeRepoPebble.Path)
		if err != nil {
			return fmt.Errorf("open node repo: %w", err)
		}

		name := []byte(args[0])
		height := math.MaxInt32

		if len(args) == 2 {
			height, err = strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid args")
			}
		}

		nm, err := node.NewBaseManager(repo)
		if err != nil {
			return fmt.Errorf("create node manager: %w", err)
		}
		nm = node.NewNormalizingManager(nm)

		_, err = nm.IncrementHeightTo(int32(height))
		if err != nil {
			return fmt.Errorf("increment height: %w", err)
		}

		n, err := nm.Node(name)
		if err != nil || n == nil {
			return fmt.Errorf("get node: %w", err)
		}

		showNode(n)
		return nil
	},
}
