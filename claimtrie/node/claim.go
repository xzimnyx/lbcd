package node

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/claimtrie/param"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

// ClaimID represents a Claim's ClaimID.
type ClaimID [20]byte

// NewClaimID returns a Claim ID caclculated from Ripemd160(Sha256(OUTPOINT).
func NewClaimID(op wire.OutPoint) ClaimID {

	w := bytes.NewBuffer(op.Hash[:])
	if err := binary.Write(w, binary.BigEndian, op.Index); err != nil {
		panic(err)
	}
	var id ClaimID
	copy(id[:], btcutil.Hash160(w.Bytes()))

	return id
}

// NewIDFromString returns a Claim ID from a string.
func NewIDFromString(s string) (ClaimID, error) {

	var id ClaimID
	_, err := hex.Decode(id[:], []byte(s))
	for i, j := 0, len(id)-1; i < j; i, j = i+1, j-1 {
		id[i], id[j] = id[j], id[i]
	}

	return id, err
}

func (id ClaimID) String() string {

	for i, j := 0, len(id)-1; i < j; i, j = i+1, j-1 {
		id[i], id[j] = id[j], id[i]
	}

	return hex.EncodeToString(id[:])
}

type Status int

const (
	Added Status = iota
	Deleted
	Accepted
	Activated
)

// Claim defines a structure of a Claim (or Support).
type Claim struct {
	OutPoint   wire.OutPoint
	ClaimID    string
	Amount     int64
	AcceptedAt int32
	ActiveAt   int32
	Status     Status

	Value []byte // Claim Only
}

func (c *Claim) SetOutPoint(op wire.OutPoint) *Claim { c.OutPoint = op; return c }
func (c *Claim) SetID(id string) *Claim              { c.ClaimID = id; return c }
func (c *Claim) SetAmt(amt int64) *Claim             { c.Amount = amt; return c }
func (c *Claim) SetAccepted(ht int32) *Claim         { c.AcceptedAt = ht; return c }
func (c *Claim) SetActiveAt(ht int32) *Claim         { c.ActiveAt = ht; return c }
func (c *Claim) SetValue(val []byte) *Claim          { c.Value = val; return c }
func (c *Claim) SetStatus(st Status) *Claim          { c.Status = st; return c }

func (c *Claim) TotalAmount(supports list) int64 {

	if c.Status != Activated && c.Status != Added {
		return 0
	}

	amt := c.Amount

	for _, s := range supports {
		if s.ClaimID == c.ClaimID && s.Status == Added {
			amt += s.Amount
		}
	}

	return amt
}

func (c *Claim) EffectiveAmount(supports list) int64 {

	if c.Status != Activated {
		return 0
	}

	amt := c.Amount

	for _, s := range supports {
		if s.ClaimID == c.ClaimID && s.Status == Activated {
			amt += s.Amount
		}
	}

	return amt
}

func (c *Claim) ExpireAt() int32 {

	if c.AcceptedAt+param.OriginalClaimExpirationTime > param.ExtendedClaimExpirationForkHeight {
		return c.AcceptedAt + param.ExtendedClaimExpirationTime
	}

	return c.AcceptedAt + param.OriginalClaimExpirationTime
}

func equal(a, b *Claim) bool {

	if a != nil && b != nil {
		return a.OutPoint == b.OutPoint
	}

	return a == nil && b == nil
}

func OutPointLess(a, b wire.OutPoint) bool {

	switch cmp := bytes.Compare(a.Hash[:], b.Hash[:]); {
	case cmp > 0:
		return true
	case cmp < 0:
		return false
	default:
		return a.Index < b.Index
	}
}

func NewOutPointFromString(str string) *wire.OutPoint {

	f := strings.Split(str, ":")
	hash, _ := chainhash.NewHashFromStr(f[0])
	idx, _ := strconv.Atoi(f[1])

	return wire.NewOutPoint(hash, uint32(idx))
}
