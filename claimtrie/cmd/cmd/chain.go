package cmd

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"

	"github.com/btcsuite/btcd/claimtrie"
	"github.com/btcsuite/btcd/claimtrie/block"
	"github.com/btcsuite/btcd/claimtrie/block/blockrepo"
	"github.com/btcsuite/btcd/claimtrie/chain/chainrepo"
	"github.com/btcsuite/btcd/claimtrie/change"
	"github.com/btcsuite/btcd/claimtrie/config"
	"github.com/cockroachdb/pebble"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(chainCmd)

	chainCmd.AddCommand(chainDumpCmd)
	chainCmd.AddCommand(chainReplayCmd)
}

var chainCmd = &cobra.Command{
	Use:   "chain",
	Short: "chain related command",
}

var chainDumpCmd = &cobra.Command{
	Use:   "dump  <fromHeight> [<toHeight>]",
	Short: "dump changes from <fromHeight> to [<toHeight>]",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {

		fromHeight, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid args")
		}

		toHeight := fromHeight + 1
		if len(args) == 2 {
			toHeight, err = strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid args")
			}
		}

		chainRepo, err := chainrepo.NewPebble(filepath.Join(cfg.DataDir, cfg.ChainRepoPebble.Path))
		if err != nil {
			return fmt.Errorf("open node repo: %w", err)
		}

		for height := fromHeight; height < toHeight; height++ {
			changes, err := chainRepo.Load(int32(height))
			if err == pebble.ErrNotFound {
				continue
			}
			if err != nil {
				return fmt.Errorf("load commands: %w", err)
			}

			for _, chg := range changes {
				if int(chg.Height) > height {
					break
				}
				showChange(chg)
			}
		}

		return nil
	},
}

var chainReplayCmd = &cobra.Command{
	Use:   "replay <height>",
	Short: "Replay the chain up to <height>",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {

		fmt.Printf("not working until we pass record flag to claimtrie\n")

		fromHeight := 2
		toHeight := int(math.MaxInt32)

		var err error
		if len(args) == 1 {
			toHeight, err = strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid args")
			}
		}

		err = os.RemoveAll(filepath.Join(cfg.DataDir, cfg.NodeRepoPebble.Path))
		if err != nil {
			return fmt.Errorf("delete node repo: %w", err)
		}

		fmt.Printf("Deleted node repo\n")

		chainRepo, err := chainrepo.NewPebble(filepath.Join(cfg.DataDir, cfg.ChainRepoPebble.Path))
		if err != nil {
			return fmt.Errorf("open change repo: %w", err)
		}

		reportedBlockRepo, err := blockrepo.NewPebble(filepath.Join(cfg.DataDir, cfg.ReportedBlockRepoPebble.Path))
		if err != nil {
			return fmt.Errorf("open block repo: %w", err)
		}

		cfg := config.DefaultConfig
		ct, err := claimtrie.New(cfg)
		if err != nil {
			return fmt.Errorf("create claimtrie: %w", err)
		}
		defer ct.Close()

		err = ct.ResetHeight(int32(fromHeight - 1))
		if err != nil {
			return fmt.Errorf("reset claimtrie height: %w", err)
		}

		for height := int32(fromHeight); height < int32(toHeight); height++ {

			changes, err := chainRepo.Load(height)
			if err == pebble.ErrNotFound {
				// do nothing.
			} else if err != nil {
				return fmt.Errorf("load from change repo: %w", err)
			}

			for _, chg := range changes {

				switch chg.Type {
				case change.AddClaim:
					err = ct.AddClaim(chg.Name, chg.OutPoint, chg.ClaimID, chg.Amount)

				case change.UpdateClaim:
					err = ct.UpdateClaim(chg.Name, chg.OutPoint, chg.Amount, chg.ClaimID)

				case change.SpendClaim:
					err = ct.SpendClaim(chg.Name, chg.OutPoint, chg.ClaimID)

				case change.AddSupport:
					err = ct.AddSupport(chg.Name, chg.OutPoint, chg.Amount, chg.ClaimID)

				case change.SpendSupport:
					err = ct.SpendSupport(chg.Name, chg.OutPoint, chg.ClaimID)

				default:
					err = fmt.Errorf("invalid change: %v", chg)
				}

				if err != nil {
					return fmt.Errorf("execute change %v: %w", chg, err)
				}
			}
			err = appendBlock(ct, reportedBlockRepo)
			if err != nil {
				return err
			}
			if ct.Height()%1000 == 0 {
				fmt.Printf("block: %d\n", ct.Height())
			}
		}

		return nil
	},
}

func appendBlock(ct *claimtrie.ClaimTrie, blockRepo block.Repo) error {

	err := ct.AppendBlock()
	if err != nil {
		return fmt.Errorf("append block: %w", err)
	}

	height := ct.Height()

	hash, err := blockRepo.Get(height)
	if err != nil {
		return fmt.Errorf("load from block repo: %w", err)
	}

	if *ct.MerkleHash() != *hash {
		return fmt.Errorf("hash mismatched at height %5d: exp: %s, got: %s", height, hash, ct.MerkleHash())
	}

	return nil
}
