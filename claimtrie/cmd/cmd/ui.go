package cmd

import (
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/claimtrie/change"
	"github.com/btcsuite/btcd/claimtrie/node"
)

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

func showChange(chg change.Change) {
	fmt.Printf(">>> Height: %6d: %s\n", chg.Height, changeName(chg.Type))
}

func showClaim(c *node.Claim, n *node.Node) {
	mark := " "
	if c == n.BestClaim {
		mark = "*"
	}

	fmt.Printf("%s  C  id: %s, op: %s, %5d/%-5d, %9s, amt: %15d, eff: %15d\n",
		mark, c.ClaimID, c.OutPoint, c.AcceptedAt, c.ActiveAt, status[c.Status], c.Amount, c.EffectiveAmount(n.Supports))
}

func showSupport(c *node.Claim) {
	fmt.Printf("    S id: %s, op: %s, %5d/%-5d, %9s, amt: %15d\n",
		c.ClaimID, c.OutPoint, c.AcceptedAt, c.ActiveAt, status[c.Status], c.Amount)
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
