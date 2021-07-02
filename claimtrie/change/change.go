package change

type ChangeType int

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

	Name     []byte
	ClaimID  string
	OutPoint string
	Amount   int64
	Value    []byte

	ActiveHeight  int32 // for normalization fork
	VisibleHeight int32
}

func New(typ ChangeType) Change {
	return Change{Type: typ}
}

func (c Change) SetHeight(height int32) Change {
	c.Height = height
	return c
}

func (c Change) SetName(name []byte) Change {
	c.Name = name
	return c
}

func (c Change) SetClaimID(claimID string) Change {
	c.ClaimID = claimID
	return c
}

func (c Change) SetOutPoint(op string) Change {
	c.OutPoint = op
	return c
}

func (c Change) SetAmount(amt int64) Change {
	c.Amount = amt
	return c
}

func (c Change) SetValue(value []byte) Change {
	c.Value = value
	return c
}
