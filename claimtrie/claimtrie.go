package claimtrie

import (
	"fmt"

	"github.com/btcsuite/btcd/claimtrie/block"
	"github.com/btcsuite/btcd/claimtrie/block/blockrepo"
	"github.com/btcsuite/btcd/claimtrie/chain"
	"github.com/btcsuite/btcd/claimtrie/chain/chainrepo"
	"github.com/btcsuite/btcd/claimtrie/change"
	"github.com/btcsuite/btcd/claimtrie/config"
	"github.com/btcsuite/btcd/claimtrie/merkletrie"
	"github.com/btcsuite/btcd/claimtrie/merkletrie/merkletrierepo"
	"github.com/btcsuite/btcd/claimtrie/node"
	"github.com/btcsuite/btcd/claimtrie/node/noderepo"
	"github.com/btcsuite/btcd/claimtrie/temporal"
	"github.com/btcsuite/btcd/claimtrie/temporal/temporalrepo"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

// ClaimTrie implements a Merkle Trie supporting linear history of commits.
type ClaimTrie struct {

	// Repository for reported block hashes (debugging purpose).
	reportedBlockRepo block.Repo

	// Repository for raw changes recieved from chain.
	chainRepo chain.Repo

	// Repository for calculated block hashes.
	blockRepo block.Repo

	// Repository for storing temporal information of nodes at each block height.
	// For example, which nodes (by name) should be refreshed at each block height
	// due to stake expiration or delayed activation.
	temporalRepo temporal.Repo

	// Cache layer of Nodes.
	nodeManager *node.Manager

	// Prefix tree (trie) that manages merkle hash of each node.
	merkleTrie *merkletrie.MerkleTrie

	// Current block height, which is increased by one when AppendBlock() is called.
	height int32

	// Write buffer for batching changes written to repo.
	// flushed before block is appended.
	changes []change.Change

	// Registrered cleanup functions which are invoked in the Close() in reverse order.
	cleanups []func() error
}

func New(record bool) (*ClaimTrie, error) {

	cfg := config.Config
	var cleanups []func() error

	blockRepo, err := blockrepo.NewPebble(cfg.BlockRepoPebble.Path)
	if err != nil {
		return nil, fmt.Errorf("new block repo: %w", err)
	}
	cleanups = append(cleanups, blockRepo.Close)

	temporalRepo, err := temporalrepo.NewPebble(cfg.TemporalRepoPebble.Path)
	if err != nil {
		return nil, fmt.Errorf("new temporal repo: %w", err)
	}
	cleanups = append(cleanups, temporalRepo.Close)

	// Initialize repository for changes to nodes.
	// The cleanup is delegated to the Node Manager.
	nodeRepo, err := noderepo.NewPebble(cfg.NodeRepoPebble.Path)
	if err != nil {
		return nil, fmt.Errorf("new node repo: %w", err)
	}

	nodeManager, err := node.NewManager(nodeRepo)
	if err != nil {
		return nil, fmt.Errorf("new node manager: %w", err)
	}
	cleanups = append(cleanups, nodeManager.Close)

	// Initialize repository for MerkleTrie.
	// The cleanup is delegated to MerkleTrie.
	trieRepo, err := merkletrierepo.NewPebble(cfg.MerkleTrieRepoPebble.Path)
	if err != nil {
		return nil, fmt.Errorf("new trie repo: %w", err)
	}

	trie := merkletrie.New(nodeManager, trieRepo)
	cleanups = append(cleanups, trie.Close)

	// Restore the last height.
	previousHeight, err := blockRepo.Load()
	if err != nil {
		return nil, fmt.Errorf("load blocks: %w", err)
	}

	// If the last height is not 0, restore the root trie node.
	if previousHeight != 0 {
		hash, err := blockRepo.Get(previousHeight)
		if err != nil {
			return nil, fmt.Errorf("get hash: %w", err)
		}
		trie.SetRoot(hash)
	}

	reportedBlockRepo, err := blockrepo.NewPebble(cfg.ReportedBlockRepoPebble.Path)
	if err != nil {
		return nil, fmt.Errorf("new reported block repo: %w", err)
	}
	cleanups = append(cleanups, reportedBlockRepo.Close)

	chainRepo, err := chainrepo.NewPebble(cfg.ChainRepoPebble.Path)
	if err != nil {
		return nil, fmt.Errorf("new change change repo: %w", err)
	}
	cleanups = append(cleanups, chainRepo.Close)

	ct := &ClaimTrie{
		blockRepo:    blockRepo,
		temporalRepo: temporalRepo,

		nodeManager: nodeManager,
		merkleTrie:  trie,

		height: previousHeight,

		reportedBlockRepo: reportedBlockRepo,
		chainRepo:         chainRepo,

		cleanups: cleanups,
	}

	return ct, nil
}

// AddClaim adds a Claim to the ClaimTrie.
func (ct *ClaimTrie) AddClaim(name []byte, op wire.OutPoint, amt int64, val []byte) error {

	chg := change.Change{
		Type:     change.AddClaim,
		Name:     name,
		OutPoint: op.String(),
		Amount:   amt,
		ClaimID:  node.NewClaimID(op).String(),
		Value:    val,
	}

	return ct.forwardNodeChange(chg)
}

// UpdateClaim updates a Claim in the ClaimTrie.
func (ct *ClaimTrie) UpdateClaim(name []byte, op wire.OutPoint, amt int64, id node.ClaimID, val []byte) error {

	chg := change.Change{
		Type:     change.UpdateClaim,
		Name:     name,
		OutPoint: op.String(),
		Amount:   amt,
		ClaimID:  id.String(),
		Value:    val,
	}

	return ct.forwardNodeChange(chg)
}

// SpendClaim spends a Claim in the ClaimTrie.
func (ct *ClaimTrie) SpendClaim(name []byte, op wire.OutPoint) error {

	chg := change.Change{
		Type:     change.SpendClaim,
		Name:     name,
		OutPoint: op.String(),
	}

	return ct.forwardNodeChange(chg)
}

// AddSupport adds a Support to the ClaimTrie.
func (ct *ClaimTrie) AddSupport(name []byte, op wire.OutPoint, amt int64, id node.ClaimID) error {

	chg := change.Change{
		Type:     change.AddSupport,
		Name:     name,
		OutPoint: op.String(),
		Amount:   amt,
		ClaimID:  id.String(),
	}

	return ct.forwardNodeChange(chg)
}

// SpendSupport spends a Support in the ClaimTrie.
func (ct *ClaimTrie) SpendSupport(name []byte, op wire.OutPoint) error {

	chg := change.Change{
		Type:     change.SpendSupport,
		Name:     name,
		OutPoint: op.String(),
	}

	return ct.forwardNodeChange(chg)
}

// AppendBlock increases block by one.
func (ct *ClaimTrie) AppendBlock() error {

	if len(ct.changes) > 0 && ct.chainRepo != nil {
		err := ct.chainRepo.Save(ct.changes)
		if err != nil {
			return fmt.Errorf("chain change repo save: %w", err)
		}
		// Truncate the buffer to zero.
		ct.changes = ct.changes[:0]
	}

	ct.height++
	ct.nodeManager.IncrementHeightTo(ct.height)

	names, err := ct.temporalRepo.NodesAt(ct.height)
	if err != nil {
		return fmt.Errorf("temporal repo nodes at: %w", err)
	}

	for _, name := range names {

		ct.merkleTrie.Update(name)

		nextupdateHeight, err := ct.nodeManager.NextUpdateHeightOfNode(name)
		if err != nil {
			return fmt.Errorf("temporal repo nodes at: %w", err)
		}

		ct.temporalRepo.SetNodeAt(name, nextupdateHeight)
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

// ResetHeight resets the ClaimTrie to a previous known height..
func (ct *ClaimTrie) ResetHeight(height int32) error {

	// TODO
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
			return fmt.Errorf("cleanup: %w", err)
		}
	}

	return nil
}

func (ct *ClaimTrie) forwardNodeChange(chg change.Change) error {

	chg.Height = ct.Height() + 1

	err := ct.nodeManager.AppendChange(chg)
	if err != nil {
		return fmt.Errorf("node manager handle change: %w", err)
	}

	err = ct.temporalRepo.SetNodeAt(chg.Name, ct.height+1)
	if err != nil {
		return fmt.Errorf("set temporal node: %w", err)
	}

	ct.changes = append(ct.changes, chg)

	return nil
}
