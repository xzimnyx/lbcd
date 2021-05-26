// Copyright (c) 2021 - LBRY Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/claimtrie/config"
	"github.com/btcsuite/btcd/claimtrie/node"
	"github.com/btcsuite/btcd/claimtrie/repo"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(nodeCmd)
}

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "replay node commands",
	RunE:  playNode,
}

func playNode(cmd *cobra.Command, args []string) error {

	nodeRepo, err := repo.NewNodeRepoPostgres(config.Config.NodeRepo.DSN)
	if err != nil {
		return fmt.Errorf("open node repo: %w", err)
	}

	changes, err := nodeRepo.Load("one", 30573)
	if err != nil {
		return fmt.Errorf("load commands: %w", err)
	}

	n := node.NewNode()
	var lastHeight int32
	for _, chg := range changes {
		chg.Value = nil
		if err := n.HandleChange(chg); err != nil {
			return fmt.Errorf("apply change: %w", err)
		}
		fmt.Printf("\n%10s Height: %5d\n\n", chg.Type, chg.Height)
		showNode(n)
		if lastHeight != chg.Height {
			n.AdjustTo(lastHeight)
			fmt.Printf("\n%s\n", strings.Repeat("#", 180))
		}
		lastHeight = chg.Height
	}

	n.AdjustTo(lastHeight)
	fmt.Printf("\n%s\n", strings.Repeat("#", 180))
	showNode(n)
	fmt.Printf("succedded\n")

	return nil
}

var status = map[node.Status]string{
	node.Added:     "Added",
	node.Accepted:  "Accepted",
	node.Activated: "Activated",
	node.Deleted:   "Deleted",
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

	fmt.Printf("  N Height: %d, Tookover: %d\n\n", n.Height, n.TakenOverAt)
	for _, c := range n.Claims {
		showClaim(c, n)
		for _, s := range n.Supports {
			if s.ClaimID != c.ClaimID {
				continue
			}
			showSupport(s)
		}
	}
}
