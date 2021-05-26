package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "claimtrie",
	Short:        "ClaimTrie Command Line Interface",
	SilenceUsage: true,
}

func Execute() {
	rootCmd.Execute() // nolint : errchk
}
