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
	Accepted Status = iota
	Activated
	Deactivated
)

// Claim defines a structure of stake, which could be a Claim or Support.
type Claim struct {
	OutPoint   wire.OutPoint
	ClaimID    string
	Amount     int64
	AcceptedAt int32 // when arrived (aka, originally landed in block)
	ActiveAt   int32 // AcceptedAt + actual delay
	Status     Status
	Value      []byte
	VisibleAt  int32
}

func (c *Claim) setOutPoint(op wire.OutPoint) *Claim {
	c.OutPoint = op
	return c
}

func (c *Claim) SetAmt(amt int64) *Claim {
	c.Amount = amt
	return c
}

func (c *Claim) setAccepted(height int32) *Claim {
	c.AcceptedAt = height
	return c
}

func (c *Claim) setActiveAt(height int32) *Claim {
	c.ActiveAt = height
	return c
}

func (c *Claim) SetValue(value []byte) *Claim {
	c.Value = value
	return c
}

func (c *Claim) setStatus(status Status) *Claim {
	c.Status = status
	return c
}

func (c *Claim) EffectiveAmount(supports ClaimList) int64 {

	if c.Status != Activated {
		return 0
	}

	amt := c.Amount

	for _, s := range supports {
		if s.Status == Activated && s.ClaimID == c.ClaimID { // TODO: this comparison is hit a lot; byte comparison instead of hex would be faster
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

func OutPointLess(a, b wire.OutPoint) bool {

	switch cmp := bytes.Compare(a.Hash[:], b.Hash[:]); {
	case cmp < 0:
		return true
	case cmp > 0:
		return false
	default:
		return a.Index < b.Index
	}
}

func NewOutPointFromString(str string) *wire.OutPoint {

	f := strings.Split(str, ":")
	if len(f) != 2 {
		return nil
	}
	hash, _ := chainhash.NewHashFromStr(f[0])
	idx, _ := strconv.Atoi(f[1])

	return wire.NewOutPoint(hash, uint32(idx))
}
