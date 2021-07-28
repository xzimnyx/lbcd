package change

import (
	"github.com/btcsuite/btcd/wire"
	"github.com/mojura/enkodo"
)

type ChangeType int32

const (
	AddClaim ChangeType = iota
	SpendClaim
	UpdateClaim
	AddSupport
	SpendSupport
)

type Change struct {
	Type   ChangeType
	Height int32

	Name     []byte `msg:"-"`
	ClaimID  ClaimID
	OutPoint wire.OutPoint
	Amount   int64

	ActiveHeight  int32 // for normalization fork
	VisibleHeight int32

	SpentChildren map[string]bool
}

func NewChange(typ ChangeType) Change {
	return Change{Type: typ}
}

func (c Change) SetHeight(height int32) Change {
	c.Height = height
	return c
}

func (c Change) SetName(name []byte) Change {
	c.Name = name // need to clone it?
	return c
}

func (c Change) SetOutPoint(op *wire.OutPoint) Change {
	c.OutPoint = *op
	return c
}

func (c Change) SetAmount(amt int64) Change {
	c.Amount = amt
	return c
}

func (c *Change) MarshalEnkodo(enc *enkodo.Encoder) error {
	enc.Bytes(c.ClaimID[:])
	enc.Bytes(c.OutPoint.Hash[:])
	enc.Uint32(c.OutPoint.Index)
	enc.Int32(int32(c.Type))
	enc.Int32(c.Height)
	enc.Int32(c.ActiveHeight)
	enc.Int32(c.VisibleHeight)
	enc.Int64(c.Amount)
	if c.SpentChildren != nil {
		enc.Int32(int32(len(c.SpentChildren)))
		for key := range c.SpentChildren {
			enc.String(key)
		}
	} else {
		enc.Int32(0)
	}
	return nil
}

func (c *Change) UnmarshalEnkodo(dec *enkodo.Decoder) error {
	id := c.ClaimID[:]
	err := dec.Bytes(&id)
	op := c.OutPoint.Hash[:]
	err = dec.Bytes(&op)
	c.OutPoint.Index, err = dec.Uint32()
	t, err := dec.Int32()
	c.Type = ChangeType(t)
	c.Height, err = dec.Int32()
	c.ActiveHeight, err = dec.Int32()
	c.VisibleHeight, err = dec.Int32()
	c.Amount, err = dec.Int64()
	keys, err := dec.Int32()
	if keys > 0 {
		c.SpentChildren = map[string]bool{}
	}
	for keys > 0 {
		keys--
		key, _ := dec.String()
		c.SpentChildren[key] = true
	}
	return err
}
