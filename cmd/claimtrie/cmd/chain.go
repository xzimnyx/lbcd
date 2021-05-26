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

	"github.com/btcsuite/btcd/claimtrie"
	"github.com/btcsuite/btcd/claimtrie/block"
	"github.com/btcsuite/btcd/claimtrie/change"
	"github.com/btcsuite/btcd/claimtrie/config"
	"github.com/btcsuite/btcd/claimtrie/node"
	"github.com/btcsuite/btcd/claimtrie/repo"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(chainCmd)
}

var chainCmd = &cobra.Command{
	Use:   "chain",
	Short: "chain related command",
	RunE:  replayChain,
}

func replayChain(cmd *cobra.Command, args []string) error {

	ct, err := claimtrie.New(false)
	if err != nil {
		return fmt.Errorf("create claimtrie: %w", err)
	}
	defer ct.Close()

	cfg := config.Config
	chgRepo, err := repo.NewChangeRepoPostgres(cfg.ChangeRepo.DSN, false)
	if err != nil {
		return fmt.Errorf("open change repo: %w", err)
	}

	blkRepo, err := repo.NewBlockRepoPebble(cfg.BlockRepo.Path)
	if err != nil {
		return fmt.Errorf("open block repo: %w", err)
	}

	for {
		chg, err := chgRepo.Load()
		if err != nil {
			return fmt.Errorf("load from change repo: %w", err)
		}

		if chg.Height != ct.Height() {
			err = appendBlock(ct, blkRepo)
			if err != nil {
				return err
			}
			if ct.Height()%1000 == 0 {
				fmt.Printf("\rblock: %d", ct.Height())
			}
		}

		name := string(chg.Name)

		switch chg.Type {
		case change.AddClaim:
			op := *node.NewOutPointFromString(chg.OutPoint)
			err = ct.AddClaim(name, op, chg.Amount, chg.Value)

		case change.UpdateClaim:
			op := *node.NewOutPointFromString(chg.OutPoint)
			claimID, _ := node.NewIDFromString(chg.ClaimID)
			id := node.ClaimID(claimID)
			err = ct.UpdateClaim(name, op, chg.Amount, id, chg.Value)

		case change.SpendClaim:
			op := *node.NewOutPointFromString(chg.OutPoint)
			err = ct.SpendClaim(name, op)

		case change.AddSupport:
			op := *node.NewOutPointFromString(chg.OutPoint)
			claimID, _ := node.NewIDFromString(chg.ClaimID)
			id := node.ClaimID(claimID)
			err = ct.AddSupport(name, op, chg.Amount, id)

		case change.SpendSupport:
			op := *node.NewOutPointFromString(chg.OutPoint)
			err = ct.SpendClaim(name, op)

		default:
			err = fmt.Errorf("invalid command: %v", chg)
		}

		if err != nil {
			return fmt.Errorf("execute command %v: %w", chg, err)
		}
	}

	return nil
}

func appendBlock(ct *claimtrie.ClaimTrie, blkRepo block.BlockRepo) error {

	err := ct.AppendBlock()
	if err != nil {
		return fmt.Errorf("append block: %w", err)

	}

	height := ct.Height()
	hash, err := blkRepo.Get(height)
	if err != nil {
		return fmt.Errorf("load from block repo: %w", err)
	}

	if ct.MerkleHash() != hash {
		return fmt.Errorf("hash mismatched at height %5d: exp: %s, got: %s", height, hash, ct.MerkleHash())
	}

	return nil
}
