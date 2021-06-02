package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/claimtrie/change"
	"github.com/btcsuite/btcd/claimtrie/config"
	"github.com/btcsuite/btcd/claimtrie/node"
	"github.com/btcsuite/btcd/claimtrie/repo"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(nodeCmd)
}

var nodeCmd = &cobra.Command{
	Use:   "node <nodename> <height>",
	Short: "replay node commands",
	RunE:  replayNode,
}

func replayNode(cmd *cobra.Command, args []string) error {

	nodeChangeRepo, err := repo.NewNodeChangeRepoPebble(config.Config.ChainChangeRepoPebble.Path)
	if err != nil {
		return fmt.Errorf("open node repo: %w", err)
	}

	if len(args) != 2 {
		return fmt.Errorf("invalid args")
	}
	nodename := string(args[0])
	height, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid args")
	}

	changes, err := nodeChangeRepo.LoadByNameUpToHeight(nodename, int32(height))
	if err != nil {
		return fmt.Errorf("load commands: %w", err)
	}

	n := node.NewNode()

	for _, chg := range changes {

		if n.Height+1 != chg.Height {
			n.AdjustTo(n.Height + 1)
			showNode(n)
			if n.Height != chg.Height {
				n.AdjustTo(chg.Height - 1)
				showNode(n)
			}
		}
		fmt.Printf(">>> Height: %6d: %s\n", chg.Height, changeName(chg.Type))

		if err := n.HandleChange(chg); err != nil {
			return fmt.Errorf("apply change: %w", err)
		}
	}

	n.AdjustTo(n.Height + 1)
	showNode(n)

	return nil
}

var status = map[node.Status]string{
	node.Added:     "Added",
	node.Accepted:  "Accepted",
	node.Activated: "Activated",
	node.Deleted:   "Deleted",
}

func changeName(c change.ChangeType) string {
	switch c {
	case change.AddClaim:
		return "Addclaim"
	case change.SpendClaim:
		return "SpendClaim"
	case change.UpdateClaim:
		return "UpdateClaim"
	case change.AddSupport:
		return "Addsupport"
	case change.SpendSupport:
		return "SpendSupport"
	}
	return "Unknown"
}

func showClaim(c *node.Claim, n *node.Node) {
	mark := " "
	if c == n.BestClaim {
		mark = "*"
	}

	fmt.Printf("%s  C  id: %s, op: %s, %5d/%5d, %9s, amt: %15d, eff: %15d\n",
		mark, c.ClaimID, c.OutPoint, c.AcceptedAt, c.AcceptedAt, status[c.Status], c.Amount, c.EffectiveAmount(n.Supports))
}

func showSupport(c *node.Claim) {
	fmt.Printf("    S id: %s, op: %s, %5d/%5d, %9s, amt: %15d\n",
		c.ClaimID, c.OutPoint, c.AcceptedAt, c.AcceptedAt, status[c.Status], c.Amount)
}

func showNode(n *node.Node) {

	fmt.Printf("%s\n", strings.Repeat("-", 200))
	fmt.Printf("Block Height: %d, Tookover: %d\n\n", n.Height, n.TakenOverAt)
	for _, c := range n.Claims {
		showClaim(c, n)
		for _, s := range n.Supports {
			if s.ClaimID != c.ClaimID {
				continue
			}
			showSupport(s)
		}
	}
	fmt.Printf("\n\n")
}
