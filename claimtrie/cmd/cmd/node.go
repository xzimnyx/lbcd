package cmd

import (
	"fmt"
	"math"
	"strconv"

	"github.com/btcsuite/btcd/claimtrie/config"
	"github.com/btcsuite/btcd/claimtrie/node"
	"github.com/btcsuite/btcd/claimtrie/node/noderepo"
	"github.com/btcsuite/btcd/claimtrie/param"
	"github.com/btcsuite/btcd/wire"

	"github.com/spf13/cobra"
)

func init() {
	param.SetNetwork(wire.MainNet, "mainnet")
	localConfig = config.GenerateConfig(param.ClaimtrieDataFolder)
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

		changes, err := repo.LoadChanges([]byte(name), int32(height))
		if err != nil {
			return fmt.Errorf("load commands: %w", err)
		}

		for _, chg := range changes {
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

		changes, err := repo.LoadChanges(name, int32(height))
		if err != nil {
			return fmt.Errorf("load commands: %w", err)
		}

		nm, err := node.NewManager(repo)
		if err != nil {
			return fmt.Errorf("create node manager: %w", err)
		}

		adjustNodeTo := func(height int32) error {

			err = nm.IncrementHeightTo(height)
			if err != nil {
				return fmt.Errorf("increment height: %w", err)
			}

			n, err := nm.Node(name)
			if err != nil {
				return fmt.Errorf("get node: %w", err)
			}

			showNode(n)

			return nil
		}

		for _, chg := range changes {

			if nm.Height()+1 != chg.Height {

				err = adjustNodeTo(chg.Height - 1)
				if err != nil {
					return fmt.Errorf("adjust node: %w", err)
				}
			}

			showChange(chg)

			err = nm.AppendChange(chg)
			if err != nil {
				return fmt.Errorf("append change: %w", err)
			}
		}

		err = adjustNodeTo(nm.Height() + 1)
		if err != nil {
			return fmt.Errorf("adjust node: %w", err)
		}

		return nil
	},
}
