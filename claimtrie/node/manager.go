package node

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/claimtrie/change"
	"github.com/btcsuite/btcd/claimtrie/param"
	"github.com/btcsuite/btcd/wire"
)

type Manager interface {
	AppendChange(chg change.Change) error
	IncrementHeightTo(height int32) ([][]byte, error)
	DecrementHeightTo(affectedNames [][]byte, height int32) error
	Height() int32
	Close() error
	Node(name []byte) (*Node, error)
	NextUpdateHeightOfNode(name []byte) ([]byte, int32)
	IterateNames(predicate func(name []byte) bool)
	ClaimHashes(name []byte) []*chainhash.Hash
	Hash(name []byte) *chainhash.Hash
}

type BaseManager struct {
	repo Repo

	height  int32
	cache   map[string]*Node
	changes []change.Change
}

func NewBaseManager(repo Repo) (Manager, error) {

	nm := &BaseManager{
		repo:  repo,
		cache: map[string]*Node{},
	}

	return nm, nil
}

// Node returns a node at the current height.
// The returned node may have pending changes.
func (nm *BaseManager) Node(name []byte) (*Node, error) {

	nameStr := string(name)
	n, ok := nm.cache[nameStr]
	if ok && n != nil {
		return n.AdjustTo(nm.height, -1, name), nil
	}

	changes, err := nm.repo.LoadChanges(name)
	if err != nil {
		return nil, fmt.Errorf("load changes from node repo: %w", err)
	}

	n, err = nm.newNodeFromChanges(changes, nm.height)
	if err != nil {
		return nil, fmt.Errorf("create node from changes: %w", err)
	}

	if n == nil { // they've requested a nonexistent or expired name
		return nil, nil
	}

	nm.cache[nameStr] = n
	return n, nil
}

// newNodeFromChanges returns a new Node constructed from the changes.
// The changes must preserve their order received.
func (nm *BaseManager) newNodeFromChanges(changes []change.Change, height int32) (*Node, error) {

	if len(changes) == 0 {
		return nil, nil
	}

	n := New()
	previous := changes[0].Height

	for _, chg := range changes {
		if chg.Height < previous {
			return nil, fmt.Errorf("expected the changes to be in order by height")
		}
		if previous < chg.Height {
			n.AdjustTo(previous, chg.Height-1, chg.Name) // update bids and activation
			previous = chg.Height
		}

		delay := nm.getDelayForName(n, chg)
		err := n.ApplyChange(chg, delay)
		if err != nil {
			return nil, fmt.Errorf("append change: %w", err)
		}
	}

	lastChange := changes[len(changes)-1]
	return n.AdjustTo(lastChange.Height, height, lastChange.Name), nil
}

func (nm *BaseManager) AppendChange(chg change.Change) error {

	if len(nm.changes) <= 0 {
		// this little code block is acting as a "block complete" method
		// that could be called after the merkle hash is complete
		if len(nm.cache) > param.MaxNodeManagerCacheSize {
			// TODO: use a better cache model?
			nm.cache = map[string]*Node{}
		}
	}

	delete(nm.cache, string(chg.Name))
	nm.changes = append(nm.changes, chg)

	return nil
}

func (nm *BaseManager) IncrementHeightTo(height int32) ([][]byte, error) {

	if height <= nm.height {
		panic("invalid height")
	}

	names := make([][]byte, 0, len(nm.changes))
	for i := range nm.changes {
		names = append(names, nm.changes[i].Name)
	}

	if err := nm.repo.AppendChanges(nm.changes); err != nil {
		return nil, fmt.Errorf("save changes to node repo: %w", err)
	}

	// Truncate the buffer size to zero.
	if len(nm.changes) > 1000 { // TODO: determine a good number here
		nm.changes = nil // release the RAM
	} else {
		nm.changes = nm.changes[:0]
	}
	nm.height = height

	return names, nil
}

func (nm *BaseManager) DecrementHeightTo(affectedNames [][]byte, height int32) error {
	if height >= nm.height {
		return fmt.Errorf("invalid height")
	}

	for _, name := range affectedNames {
		delete(nm.cache, string(name))
		if err := nm.repo.DropChanges(name, height); err != nil {
			return err
		}
	}

	nm.height = height

	return nil
}

func (nm *BaseManager) getDelayForName(n *Node, chg change.Change) int32 {
	if n.BestClaim != nil {
		if n.BestClaim.ClaimID == chg.ClaimID {
			return 0
		}
	}
	if chg.ActiveHeight >= chg.Height {
		return chg.ActiveHeight - chg.Height
	}
	forceZero := false
	if chg.Height >= param.MaxRemovalWorkaroundHeight {
		if n.BestClaim == nil {
			return 0
		}

		// LEFT OFF: this code below may or may not be right. It's likely too slow of an approach.
		// We can simulate the old cache by putting each name into a map at true, and then dropping each letter and putting that in the map at false.

		// The first two to need it are:
		// Height: 664642, en-vivo-hablando-de-bitcoin-y-3
		// Height: 664642, en-vivo-hablando-de-bitcoin-y-4

		// TODO: hard fork this out; it's a bug from previous versions:
		//if nm.hasChildren(chg.Name, chg.Height, 2) {
		//	fmt.Printf("Removal Workaround affects %s at %d, %d\n", chg.Name, chg.Height, len(n.Claims))
		//	return 0
		//}
	} else if len(n.Claims) > 0 {
		// NOTE: old code had a bug in it where nodes with no claims but with children would get left in the cache after removal.
		// This would cause the getNumBlocksOfContinuousOwnership to return zero (causing incorrect takeover height calc).
		w, ok := param.DelayWorkarounds[string(chg.Name)]
		if ok {
			for _, h := range w {
				if chg.Height == h {
					forceZero = true
					break
				}
			}
		}
	}
	if n.BestClaim == nil {
		return 0
	}
	delay := calculateDelay(chg.Height, n.TakenOverAt)
	if delay > 0 && forceZero {
		return 0
	}
	if delay > 0 && !forceZero && len(n.Claims) > 0 {
		ok := param.RealRemovals[fmt.Sprintf("%d:%s", chg.Height, chg.Name)]
		if ok {
			fmt.Printf("MISSING RM: \"%s\": %d,\n", chg.Name, chg.Height)
			return 0
		}
	}
	return delay
}

func calculateDelay(curr, tookOver int32) int32 {

	delay := (curr - tookOver) / param.ActiveDelayFactor
	if delay > param.MaxActiveDelay {
		return param.MaxActiveDelay
	}

	return delay
}

func (nm BaseManager) NextUpdateHeightOfNode(name []byte) ([]byte, int32) {

	n, err := nm.Node(name)
	if err != nil || n == nil {
		return name, 0
	}

	return name, n.NextUpdate()
}

func (nm *BaseManager) Height() int32 {
	return nm.height
}

func (nm *BaseManager) Close() error {

	err := nm.repo.Close()
	if err != nil {
		return fmt.Errorf("close repo: %w", err)
	}

	return nil
}

func (nm *BaseManager) hasChildren(name []byte, height int32, required int) bool {
	c := 0
	nm.repo.IterateChildren(name, func(changes []change.Change) bool {
		// if the key is unseen, generate a node for it to height
		// if that node is active then increase the count
		n, _ := nm.newNodeFromChanges(changes, height)
		if n != nil && n.BestClaim != nil && n.BestClaim.Status == Activated {
			c++
			if c >= required {
				return false
			}
		}
		return true
	})
	return c >= required
}

func (nm *BaseManager) IterateNames(predicate func(name []byte) bool) {
	nm.repo.IterateAll(predicate)
}

func (nm *BaseManager) ClaimHashes(name []byte) []*chainhash.Hash {

	n, err := nm.Node(name)
	if err != nil {
		return nil
	}
	n.SortClaims()
	claimHashes := make([]*chainhash.Hash, 0, len(n.Claims))
	for _, c := range n.Claims {
		if c.Status == Activated { // TODO: unit test this line
			claimHashes = append(claimHashes, calculateNodeHash(c.OutPoint, n.TakenOverAt))
		}
	}
	return claimHashes
}

func (nm *BaseManager) Hash(name []byte) *chainhash.Hash {

	n, err := nm.Node(name)
	if err != nil {
		return nil
	}
	if n != nil && len(n.Claims) > 0 {
		if n.BestClaim != nil && n.BestClaim.Status == Activated {
			return calculateNodeHash(n.BestClaim.OutPoint, n.TakenOverAt)
		}
	}
	return nil
}

func calculateNodeHash(op wire.OutPoint, takeover int32) *chainhash.Hash {

	txHash := chainhash.DoubleHashH(op.Hash[:])

	nOut := []byte(strconv.Itoa(int(op.Index)))
	nOutHash := chainhash.DoubleHashH(nOut)

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(takeover))
	heightHash := chainhash.DoubleHashH(buf)

	h := make([]byte, 0, sha256.Size*3)
	h = append(h, txHash[:]...)
	h = append(h, nOutHash[:]...)
	h = append(h, heightHash[:]...)

	hh := chainhash.DoubleHashH(h)

	return &hh
}
