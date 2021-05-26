package node

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/claimtrie/change"
)

type Manager struct {
	repo Repo

	height int32
	cache  map[string]*Node

	changes []change.Change
}

func NewManager(repo Repo) (*Manager, error) {

	nm := &Manager{
		repo:  repo,
		cache: map[string]*Node{},
	}

	return nm, nil
}

// Node returns a node at the current height.
// The returned node may have pending changes.
func (nm *Manager) Node(name []byte) (*Node, error) {

	nameStr := string(name)

	n, ok := nm.cache[nameStr]
	if !ok {

		var err error
		changes, err := nm.repo.LoadChanges(name, nm.height)
		if err != nil {
			return nil, fmt.Errorf("load changes from node repo: %w", err)
		}

		n, err = NewNodeFromChanges(changes)
		if err != nil {
			return nil, fmt.Errorf("create node from changes: %w", err)
		}

		nm.cache[nameStr] = n
	}

	// With any pending changes, the node must being updated already.
	if n.pendingChanges {
		return n, nil
	}

	// The node may be at an earlier height.
	// It could be a stalled cached version or reconstructed one.
	return n.AdjustTo(nm.height), nil
}

func (nm *Manager) AppendChange(chg change.Change) error {

	nm.changes = append(nm.changes, chg)

	n, err := nm.Node(chg.Name)
	if err != nil {
		return fmt.Errorf("node manager get node: %w", err)
	}

	err = n.AppendChange(chg)
	if err != nil {
		return fmt.Errorf("handle change: %w", err)
	}

	return nil
}

func (nm *Manager) IncrementHeightTo(height int32) error {

	if height < nm.height {
		return fmt.Errorf("invalid height")
	}

	// Flush the changes to the repo.
	if height > nm.height {

		if err := nm.repo.SaveChanges(nm.changes); err != nil {
			return fmt.Errorf("save changes to node repo: %w", err)
		}

		for _, chg := range nm.changes {
			n, err := nm.Node(chg.Name)
			if err != nil {
				return fmt.Errorf("node manager get: %w", err)
			}

			if n.pendingChanges {
				n.AdjustTo(nm.height)
			}
		}

		// Truncate the buffer size to zero.
		nm.changes = nm.changes[:0]
	}

	nm.height = height

	return nil
}

func (nm Manager) NextUpdateHeightOfNode(name []byte) (int32, error) {

	n, err := nm.Node(name)
	if err != nil {
		return 0, fmt.Errorf("node manager get: %w", err)
	}

	if n.pendingChanges {
		n.AdjustTo(nm.height)
	}

	return n.NextUpdate(), nil
}

// Get returns implements merkletrie.Value interface.
// Get should only be called when all pending changes have been applied.
func (nm Manager) Get(key []byte) (*chainhash.Hash, error) {

	n, err := nm.Node(key)
	if err != nil {
		return nil, fmt.Errorf("get node: %w", err)
	}

	if n.pendingChanges {
		return nil, fmt.Errorf("pending changes should have been applied")
	}

	return n.Hash(), nil
}

func (nm *Manager) Height() int32 {
	return nm.height
}

func (nm *Manager) Close() error {

	err := nm.repo.Close()
	if err != nil {
		return fmt.Errorf("close repo: %w", err)
	}

	return nil
}
