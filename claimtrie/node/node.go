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

// ErrNotFound is returned when the Claim or Support is not found.
var ErrNotFound = fmt.Errorf("not found")

type Node struct {
	Height      int32  // Current height.
	BestClaim   *Claim // The claim that has most effective amount at the current height.
	TakenOverAt int32  // The height at when the current BestClaim Tookover.
	Claims      list   // List of all Claims.
	Supports    list   // List of all Supports, including orphaned ones.

	dirty bool
}

// NewNode returns a new Node.
func NewNode() *Node {
	return &Node{
		Claims:   list{},
		Supports: list{},
	}
}

// NewNodeFromChanges returns a new Node.
func NewNodeFromChanges(changes []change.Change) (*Node, error) {

	n := NewNode()

	var lastHeight int32
	for _, chg := range changes {
		chg.Value = nil
		if err := n.HandleChange(chg); err != nil {
			return nil, fmt.Errorf("handle change: %s", err)
		}
		if lastHeight != chg.Height {
			n.AdjustTo(lastHeight)
		}
		lastHeight = chg.Height
	}

	n.AdjustTo(lastHeight)

	return n, nil
}

func (n *Node) HandleChange(chg change.Change) error {

	op := *NewOutPointFromString(chg.OutPoint)

	switch chg.Type {
	case change.AddClaim:
		n.Claims[op] = &Claim{
			OutPoint:   op,
			Amount:     chg.Amount,
			ClaimID:    chg.ClaimID,
			AcceptedAt: chg.Height + 1,
			Value:      chg.Value,
			Status:     Added,
		}

	case change.SpendClaim:
		c, ok := n.Claims[op]
		if !ok {
			return ErrNotFound
		}
		c.SetStatus(Deleted)

	case change.UpdateClaim:
		// Find and remove the claim, which has just been spent, and placed in the removedClaims.
		c := n.Claims.find(byID(chg.ClaimID))
		if c == nil || c.Status != Deleted {
			return ErrNotFound
		}

		delete(n.Claims, c.OutPoint)
		// Keep its ID and update the rest properties.
		c.SetOutPoint(op).
			SetAmt(chg.Amount).
			SetValue(chg.Value).
			SetStatus(Accepted).
			SetAccepted(chg.Height + 1)

		n.Claims[op] = c

		if c.ActiveAt <= n.Height {
			c.SetStatus(Activated)
		}

	case change.AddSupport:
		n.Supports[op] = &Claim{
			OutPoint:   op,
			Amount:     chg.Amount,
			ClaimID:    chg.ClaimID,
			AcceptedAt: chg.Height + 1,
			Status:     Added,
		}

	case change.SpendSupport:
		s, ok := n.Supports[op]
		if !ok {
			return ErrNotFound
		}

		s.SetStatus(Deleted)
	}

	n.dirty = true

	return nil
}

func (n *Node) commit() {

	if !n.dirty {
		return
	}

	n.Claims.removeAll(byStatus(Deleted))
	n.Supports.removeAll(byStatus(Deleted))

	if n.BestClaim != nil && n.BestClaim.Status == Deleted {
		n.BestClaim, n.TakenOverAt = nil, n.Height+1
	}

	n.handleAdded()
	n.dirty = false
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
			c.SetStatus(status).SetActiveAt(activatedAt)
		}

		for _, s := range n.Supports {
			if s.ClaimID != c.ClaimID {
				continue
			}
			if s.Status != Activated {
				s.SetStatus(status).SetActiveAt(activatedAt)
			}
		}
	}
}

func (n *Node) handleExpiredAndActivated() {

	for op, c := range n.Claims {
		if c.Status == Accepted && c.ActiveAt == n.Height {
			c.SetStatus(Activated)
		}
		if c.ExpireAt() <= n.Height {
			delete(n.Claims, op)
		}
	}

	for op, s := range n.Supports {
		if s.Status == Accepted && s.ActiveAt == n.Height {
			s.SetStatus(Activated)
		}
		if s.ExpireAt() <= n.Height {
			delete(n.Supports, op)
		}
	}
}

// AdjustTo increments current height until it reaches the specified height.
func (n *Node) AdjustTo(ht int32) *Node {

	// Handle and clear all tmp claims and supports.
	for n.Height < ht {
		if n.dirty {
			n.commit()
		}
		n.Height++
		n.handleExpiredAndActivated()
		n.bid()

		n.Height = n.NextUpdate()
		if n.Height > ht {
			n.Height = ht
		}
	}
	if !n.dirty {
		n.handleExpiredAndActivated()
		n.bid()
	}

	return n
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

func (n *Node) NextUpdate() int32 {

	if n.dirty {
		return n.Height + 1
	}

	next := int32(math.MaxInt32)

	for _, c := range n.Claims {
		if ht := c.ExpireAt(); ht < next && ht > n.Height {
			next = ht
		}
		if ht := c.ActiveAt; ht < next && ht > n.Height {
			next = ht
		}
	}

	for _, s := range n.Supports {
		if ht := s.ExpireAt(); ht < next && ht > n.Height {
			next = ht
		}
		if ht := s.ActiveAt; ht < next && ht > n.Height {
			next = ht
		}
	}
	return next
}

func (n *Node) bestPrice() int64 {

	if n.BestClaim == nil {
		return 0
	}
	amt := n.BestClaim.Amount

	for _, s := range n.Supports {
		if s.ClaimID == n.BestClaim.ClaimID {
			amt += s.Amount
		}
	}

	return amt
}

func (n *Node) findCandiadte() *Claim {

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
func (n *Node) Hash() *chainhash.Hash {

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
