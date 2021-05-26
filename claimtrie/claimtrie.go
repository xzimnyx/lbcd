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

package claimtrie

import (
	"fmt"

	"github.com/btcsuite/btcd/claimtrie/block"
	"github.com/btcsuite/btcd/claimtrie/change"
	"github.com/btcsuite/btcd/claimtrie/config"
	"github.com/btcsuite/btcd/claimtrie/merkletrie"
	"github.com/btcsuite/btcd/claimtrie/repo"
	"github.com/btcsuite/btcd/claimtrie/temporal"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/claimtrie/node"
	"github.com/btcsuite/btcd/wire"
)

// ClaimTrie implements a Merkle Trie supporting linear history of commits.
type ClaimTrie struct {

	// Repository for reported block hashes (debugging purpose).
	reportedBlockRepo block.BlockRepo

	// Repository for calculated block hashes.
	blockRepo block.BlockRepo

	// Repository for raw changes recieved from chain.
	changeRepo change.ChangeRepo

	// Repository for storing temporal information of nodes at each block height.
	// For example, which nodes (by name) should be refreshed at each block height
	// due to stake expiration or delayed activation.
	temporalRepo temporal.TemporalRepo

	// Cache layer of Nodes.
	nodeManager *node.NodeManager

	// Prefix tree (trie) that manages merkle hash of each node.
	merkleTrie *merkletrie.MerkleTrie

	// Current block height, which is increased by one when AppendBlock() is called.
	height int32

	// Registrered cleanup functions which are invoked in the Close() in reverse order.
	cleanups []func() error
}

func New(record bool) (*ClaimTrie, error) {

	cfg := config.Config
	var cleanups []func() error

	reportedBlkRepo, err := repo.NewBlockRepoPebble(cfg.ReportedBlockRepo.Path)
	if err != nil {
		return nil, fmt.Errorf("new block repo: %w", err)
	}
	cleanups = append(cleanups, reportedBlkRepo.Close)

	blockRepo, err := repo.NewBlockRepoPebble(cfg.BlockRepo.Path)
	if err != nil {
		return nil, fmt.Errorf("new block repo: %w", err)
	}
	cleanups = append(cleanups, blockRepo.Close)

	changeRepo, err := repo.NewChangeRepoPostgres(cfg.ChangeRepo.DSN, cfg.ChangeRepo.Drop)
	if err != nil {
		return nil, fmt.Errorf("new change repo: %w", err)
	}
	cleanups = append(cleanups, changeRepo.Close)

	temporalRepo, err := repo.NewTemporalPebble(cfg.TemporalRepo.Path)
	if err != nil {
		return nil, fmt.Errorf("new temporal repo: %w", err)
	}
	cleanups = append(cleanups, temporalRepo.Close)

	trieRepo, err := repo.NewTrieRepoPebble(cfg.TrieRepo.Path)
	if err != nil {
		return nil, fmt.Errorf("new trie repo: %w", err)
	}
	cleanups = append(cleanups, trieRepo.Close)

	nodeRepo, err := repo.NewNodeRepoPostgres(cfg.NodeRepo.DSN)
	if err != nil {
		return nil, fmt.Errorf("new node repo: %w", err)
	}
	cleanups = append(cleanups, nodeRepo.Close)

	nodeManager, err := node.NewNodeManager(nodeRepo)
	if err != nil {
		return nil, fmt.Errorf("new nodemanager: %w", err)
	}
	cleanups = append(cleanups, nodeManager.Close)

	trie := merkletrie.New(nodeManager, trieRepo)

	ct := &ClaimTrie{
		reportedBlockRepo: reportedBlkRepo,
		blockRepo:         blockRepo,
		changeRepo:        changeRepo,
		temporalRepo:      temporalRepo,

		nodeManager: nodeManager,
		merkleTrie:  trie,

		cleanups: cleanups,
	}

	return ct, nil
}

// AddClaim adds a Claim to the ClaimTrie.
func (ct *ClaimTrie) AddClaim(name string, op wire.OutPoint, amt int64, val []byte) error {

	chg := change.Change{
		Type:     change.AddClaim,
		Name:     []byte(name),
		OutPoint: op.String(),
		Amount:   amt,
		ClaimID:  node.NewClaimID(op).String(),
		Value:    val,
	}

	return ct.handleNodeChange(chg)
}

// UpdateClaim updates a Claim in the ClaimTrie.
func (ct *ClaimTrie) UpdateClaim(name string, op wire.OutPoint, amt int64, id node.ClaimID, val []byte) error {

	chg := change.Change{
		Type:     change.UpdateClaim,
		Name:     []byte(name),
		OutPoint: op.String(),
		Amount:   amt,
		ClaimID:  id.String(),
		Value:    val,
	}

	return ct.handleNodeChange(chg)
}

// SpendClaim spends a Claim in the ClaimTrie.
func (ct *ClaimTrie) SpendClaim(name string, op wire.OutPoint) error {

	chg := change.Change{
		Type:     change.SpendClaim,
		Name:     []byte(name),
		OutPoint: op.String(),
	}

	return ct.handleNodeChange(chg)
}

// AddSupport adds a Support to the ClaimTrie.
func (ct *ClaimTrie) AddSupport(name string, op wire.OutPoint, amt int64, id node.ClaimID) error {

	chg := change.Change{
		Type:     change.AddSupport,
		Name:     []byte(name),
		OutPoint: op.String(),
		Amount:   amt,
		ClaimID:  id.String(),
	}

	return ct.handleNodeChange(chg)
}

// SpendSupport spends a Support in the ClaimTrie.
func (ct *ClaimTrie) SpendSupport(name string, op wire.OutPoint) error {

	chg := change.Change{
		Type:     change.SpendSupport,
		Name:     []byte(name),
		OutPoint: op.String(),
	}

	return ct.handleNodeChange(chg)
}

// AppendBlock increases block by one.
func (ct *ClaimTrie) AppendBlock() error {

	ct.height++
	names, err := ct.temporalRepo.NodesAt(ct.height)
	if err != nil {
		return fmt.Errorf("internal: %w", err)
	}

	for _, name := range names {
		ct.merkleTrie.Update([]byte(name))
		next := ct.nodeAt(name).NextUpdate()
		ct.temporalRepo.SetNodeAt(name, next)
	}

	h := ct.MerkleHash()
	ct.blockRepo.Set(ct.height, h)
	ct.merkleTrie.SetRoot(h)

	return nil
}

// ReportHash persists the Merkle Hash "learned and reported" by the block.
// This is for debugging purpose.
// So we can replay the trace of changes and compare calcuated and learned hash.
func (ct *ClaimTrie) ReportHash(height int32, hash chainhash.Hash) error {

	if ct.reportedBlockRepo != nil {
		return ct.reportedBlockRepo.Set(height, &hash)
	}

	return nil
}

// ReportHash reports the MerkleHash of the receieved block.
// Note: debugging purpose. and will be deprecated when the development stablized.
func (ct *ClaimTrie) ResetHeight(ht int32) error {

	// TODO
	return nil
}

func (ct *ClaimTrie) nodeAt(name string) *node.Node {

	n := ct.nodeManager.GetNode(name)

	return n.AdjustTo(ct.height)
}

func (ct *ClaimTrie) handleNodeChange(chg change.Change) error {

	chg.Height = ct.Height() + 1
	if ct.changeRepo != nil {
		if err := ct.changeRepo.Save(chg); err != nil {
			return err
		}
	}

	n := ct.nodeAt(string(chg.Name))

	if err := n.HandleChange(chg); err != nil {
		return fmt.Errorf("handle change: %w", err)
	}

	if err := ct.temporalRepo.SetNodeAt(string(chg.Name), ct.height+1); err != nil {
		return fmt.Errorf("set temporal node: %w", err)
	}

	return nil
}

// MerkleHash returns the Merkle Hash of the claimTrie.
func (ct ClaimTrie) MerkleHash() *chainhash.Hash {
	return ct.merkleTrie.MerkleHash()
}

// Height returns the current block height.
func (ct ClaimTrie) Height() int32 {
	return ct.height
}

// Close persists states.
// Any calls to the ClaimTrie after Close() being called results undefined behaviour.
func (ct *ClaimTrie) Close() error {

	for i := len(ct.cleanups) - 1; i >= 0; i-- {
		cleanup := ct.cleanups[i]
		err := cleanup()
		if err != nil {
			return err
		}
	}

	return nil
}
