package cmd

import (
	"github.com/btcsuite/btcd/claimtrie/config"
	"github.com/btcsuite/btcd/claimtrie/param"
	"github.com/btcsuite/btcd/wire"

	"github.com/spf13/cobra"
)

var cfg = config.DefaultConfig

func init() {
	param.SetNetwork(wire.MainNet)
}

var rootCmd = &cobra.Command{
	Use:          "claimtrie",
	Short:        "ClaimTrie Command Line Interface",
	SilenceUsage: true,
}

func Execute() {
	rootCmd.Execute() // nolint : errchk
}
