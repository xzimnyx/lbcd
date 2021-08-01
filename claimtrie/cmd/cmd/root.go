package cmd

import (
	"os"

	"github.com/btcsuite/btcd/claimtrie/config"
	"github.com/btcsuite/btcd/claimtrie/param"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btclog"

	"github.com/spf13/cobra"
)

var (
	log     btclog.Logger
	cfg     = config.DefaultConfig
	netName string
	dataDir string
)

var rootCmd = NewRootCommand()

func NewRootCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:          "claimtrie",
		Short:        "ClaimTrie Command Line Interface",
		SilenceUsage: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			switch netName {
			case "mainnet":
				param.SetNetwork(wire.MainNet)
			case "testnet":
				param.SetNetwork(wire.TestNet3)
			case "regtest":
				param.SetNetwork(wire.TestNet)
			}
		},
	}

	cmd.PersistentFlags().StringVar(&netName, "netname", "mainnet", "Net name")
	cmd.PersistentFlags().StringVar(&dataDir, "datadir", cfg.DataDir, "Data dir")

	return cmd
}

func Execute() {

	backendLogger := btclog.NewBackend(os.Stdout)
	defer os.Stdout.Sync()
	log = backendLogger.Logger("CMDL")
	log.SetLevel(btclog.LevelDebug)

	rootCmd.Execute() // nolint : errchk
}
