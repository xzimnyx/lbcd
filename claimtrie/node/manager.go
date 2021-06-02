package node

import (
	"fmt"

	"github.com/btcsuite/btcd/claimtrie/merkletrie"
)

type NodeManager struct {
	repo Repo

	height int32
	cache  map[string]*Node
}

func NewNodeManager(repo Repo) (*NodeManager, error) {

	nm := &NodeManager{
		repo:  repo,
		cache: map[string]*Node{},
	}

	return nm, nil
}

func (nm *NodeManager) SetHeight(height int32) {
	nm.height = height
}

func (nm *NodeManager) Get(key []byte) (merkletrie.Value, error) {
	return nm.GetNode(key)
}

func (nm *NodeManager) GetNode(key []byte) (*Node, error) {

	name := string(key)

	n, ok := nm.cache[name]
	if !ok {
		var err error
		changes, err := nm.repo.LoadByNameUpToHeight(name, nm.height)
		if err != nil {
			return nil, fmt.Errorf("load changes from nod repo: %w", err)
		}
		n, err = NewNodeFromChanges(changes)
		if err != nil {
			return nil, fmt.Errorf("create node from changes: %w", err)
		}
		nm.cache[name] = n
	}

	n.AdjustTo(nm.height)

	return n, nil
}

func (nm *NodeManager) PurgeCache() {
	nm.cache = map[string]*Node{}
}

func (nm *NodeManager) Close() error {
	return nm.repo.Close()
}
