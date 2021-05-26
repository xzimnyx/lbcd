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
	"log"
	"strconv"

	"github.com/btcsuite/btcd/claimtrie/config"
	"github.com/btcsuite/btcd/claimtrie/repo"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(blockCmd)
}

var blockCmd = &cobra.Command{
	Use:   "block FROM_HEIGHT [TO_HEIGHT]",
	Short: "Show merkle hash of block at height",
	Args:  cobra.ExactArgs(1),
	RunE:  listHash,
}

func listHash(cmd *cobra.Command, args []string) error {

	blkRepo, err := repo.NewBlockRepoPebble(config.Config.ReportedBlockRepo.Path)
	if err != nil {
		log.Fatalf("can't open reported block repo: %s", err)
	}

	from, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid args")
	}

	to := from
	if len(args) == 2 {
		to, err = strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid args")
		}
	}

	for i := from; i < to; i++ {
		hash, err := blkRepo.Get(int32(i))
		if err != nil {
			return fmt.Errorf("load changes from repo: %w", err)
		}
		fmt.Printf("blk %-7d: %s\n", i, hash.String())
	}

	return nil
}
