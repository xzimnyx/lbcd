package node

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"

	"github.com/btcsuite/btcd/claimtrie/change"
	"github.com/btcsuite/btcd/claimtrie/param"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

// ErrNotFound is returned when a claim or support is not found.
var ErrNotFound = fmt.Errorf("not found")

type Node struct {
	Height      int32  // Current height.
	BestClaim   *Claim // The claim that has most effective amount at the current height.
	TakenOverAt int32  // The height at when the current BestClaim Tookover.
	Claims      list   // List of all Claims.
	Supports    list   // List of all Supports, including orphaned ones.

	pendingChanges bool
}

// New returns a new node.
func New() *Node {
	return &Node{
		Claims:   list{},
		Supports: list{},
	}
}

// NewNodeFromChanges returns a new Node constructed from the changes.
// The changes must preserve their order receieved.
func NewNodeFromChanges(changes []change.Change) (*Node, error) {

	n := New()

	if len(changes) == 0 {
		return n, nil
	}

	// When applying a change at H height, the node has to be at H - 1 height.
	for _, chg := range changes {

		if n.Height < chg.Height-1 {
			n.AdjustTo(chg.Height - 1)
		}

		err := n.AppendChange(chg)
		if err != nil {
			return nil, fmt.Errorf("append change: %w", err)
		}
	}

	// Handle the appended changes of the last block.
	n.AdjustTo(n.Height + 1)

	return n, nil
}

func (n *Node) AppendChange(chg change.Change) error {

	op := *NewOutPointFromString(chg.OutPoint)

	switch chg.Type {
	case change.AddClaim:
		n.Claims[op] = &Claim{
			OutPoint:   op,
			Amount:     chg.Amount,
			ClaimID:    chg.ClaimID,
			AcceptedAt: chg.Height,
			Value:      chg.Value,
			Status:     Added,
		}

	case change.SpendClaim:
		c, ok := n.Claims[op]
		if !ok {
			return ErrNotFound
		}

		c.setStatus(Deleted)

	case change.UpdateClaim:
		// Find and remove the claim, which has just been spent.
		c := n.Claims.find(byID(chg.ClaimID))
		if c == nil || c.Status != Deleted {
			return ErrNotFound
		}

		// Remove the spent one from the claim list.
		delete(n.Claims, c.OutPoint)

		// Keep its ID, which was generated from the spent claim.
		// And update the rest of properties.
		c.setOutPoint(op).SetAmt(chg.Amount).SetValue(chg.Value)

		// TODO: check the bidding rules.
		c.setStatus(Accepted)

		// TODO: check the bidding rules.
		// Does it ineherit the height, or reset with new height?
		c.setAccepted(chg.Height)

		if c.ActiveAt <= n.Height {
			c.setStatus(Activated)
		}

		// Put the updated claim back to the claim list.
		n.Claims[op] = c

	case change.AddSupport:
		s := &Claim{
			OutPoint:   op,
			Amount:     chg.Amount,
			ClaimID:    chg.ClaimID,
			AcceptedAt: chg.Height,
			Status:     Added,
		}

		if n.BestClaim != nil && n.BestClaim.ClaimID == s.ClaimID {
			s.setStatus(Activated)
		}

		n.Supports[op] = s

	case change.SpendSupport:
		s, ok := n.Supports[op]
		if !ok {
			return ErrNotFound
		}

		s.setStatus(Deleted)
	}

	n.pendingChanges = true

	return nil
}

// AdjustTo increments current height until it reaches the specified height.
func (n *Node) AdjustTo(height int32) *Node {

	if n.pendingChanges {
		n.applyPendingChanges()
	}

	for n.Height < height {
		n.Height = n.NextUpdate()
		if n.Height > height {
			n.Height = height
		}
		n.handleExpiredAndActivated()
		n.bid()
	}

	return n
}

func (n *Node) applyPendingChanges() {

	n.Claims.removeAll(byStatus(Deleted))
	n.Supports.removeAll(byStatus(Deleted))

	// The current BestClaim has been deleted. A takeover is happening.
	if n.BestClaim != nil && n.BestClaim.Status == Deleted {
		n.BestClaim = nil
		// TODO: check the bidding rules.
		n.TakenOverAt = n.Height + 1
	}

	n.handleAdded()
	n.pendingChanges = false
	n.Height++
	n.handleExpiredAndActivated()
	n.bid()
}

func (n *Node) handleAdded() {

	bestPrice := n.bestPrice()
	if n.BestClaim == nil {
		n.TakenOverAt = n.Height + 1
	}

	for _, c := range n.Claims {

		if c == n.BestClaim {
			continue
		}

		status := Activated
		activatedAt := n.Height + 1

		if c.TotalAmount(n.Supports) > bestPrice {
			status = Accepted
			activatedAt = n.Height + 1 + calculateDelay(n.Height+1, n.TakenOverAt)
		}

		if c.Status != Activated {
			c.setStatus(status).setActiveAt(activatedAt)
		}

		for _, s := range n.Supports {
			if s.ClaimID != c.ClaimID {
				continue
			}
			if s.Status != Activated {
				s.setStatus(status).setActiveAt(activatedAt)
			}
		}
	}
}

func (n *Node) handleExpiredAndActivated() {

	for op, c := range n.Claims {
		if c.Status == Accepted && c.ActiveAt == n.Height {
			c.setStatus(Activated)
		}
		if c.ExpireAt() <= n.Height {
			delete(n.Claims, op)
		}
	}

	for op, s := range n.Supports {
		if s.Status == Accepted && s.ActiveAt == n.Height {
			s.setStatus(Activated)
		}
		if s.ExpireAt() <= n.Height {
			delete(n.Supports, op)
		}
	}
}

func (n *Node) bid() {

	for {
		c := n.findCandiadte()
		if equal(n.BestClaim, c) {
			break
		}
		n.BestClaim, n.TakenOverAt = c, n.Height
	}
}

// NextUpdate returns the nearest height in the future that the node should
// be refreshed due to changes of claims or supports.
func (n Node) NextUpdate() int32 {

	// The node is at height H, and apply changes for H+1.
	if n.pendingChanges {
		return n.Height + 1
	}

	next := int32(math.MaxInt32)

	for _, c := range n.Claims {
		if height := c.ExpireAt(); height > n.Height && height < next {
			next = height
		}
		if height := c.ActiveAt; height > n.Height && height < next {
			next = height
		}
	}

	for _, s := range n.Supports {
		if height := s.ExpireAt(); height > n.Height && height < next {
			next = height
		}
		if height := s.ActiveAt; height > n.Height && height < next {
			next = height
		}
	}

	return next
}

func (n Node) bestPrice() int64 {

	if n.BestClaim == nil {
		return 0
	}

	amt := n.BestClaim.Amount

	for _, s := range n.Supports {
		if s.ClaimID == n.BestClaim.ClaimID {
			if s.Status != Activated {
				panic("bug: supports for the BestClaim should always be active")
			}
			amt += s.Amount
		}
	}

	return amt
}

func (n Node) findCandiadte() *Claim {

	var c *Claim
	for _, v := range n.Claims {

		effAmountV := v.EffectiveAmount(n.Supports)

		switch {
		case v.Status != Activated:
			continue
		case c == nil:
			c = v
		case effAmountV > c.EffectiveAmount(n.Supports):
			c = v
		case effAmountV < c.EffectiveAmount(n.Supports):
			continue
		case v.AcceptedAt < c.AcceptedAt:
			c = v
		case v.AcceptedAt > c.AcceptedAt:
			continue
		case OutPointLess(c.OutPoint, v.OutPoint):
			c = v
		}
	}

	return c
}

// Hash calculates the Hash value based on the OutPoint and when it tookover.
func (n Node) Hash() *chainhash.Hash {

	if n.BestClaim == nil {
		return nil
	}

	return calculateNodeHash(n.BestClaim.OutPoint, n.TakenOverAt)
}

func calculateDelay(curr, tookover int32) int32 {

	delay := (curr - tookover) / param.ActiveDelayFactor
	if delay > param.MaxActiveDelay {
		return param.MaxActiveDelay
	}

	return delay
}

func calculateNodeHash(op wire.OutPoint, tookover int32) *chainhash.Hash {

	txHash := chainhash.DoubleHashH(op.Hash[:])

	nOut := []byte(strconv.Itoa(int(op.Index)))
	nOutHash := chainhash.DoubleHashH(nOut)

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(tookover))
	heightHash := chainhash.DoubleHashH(buf)

	h := make([]byte, 0, sha256.Size*3)
	h = append(h, txHash[:]...)
	h = append(h, nOutHash[:]...)
	h = append(h, heightHash[:]...)

	hh := chainhash.DoubleHashH(h)

	return &hh
}
