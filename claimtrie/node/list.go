package node

import "github.com/btcsuite/btcd/wire"

type list map[wire.OutPoint]*Claim

type comparator func(c *Claim) bool

func byID(id string) comparator {
	return func(c *Claim) bool {
		return c.ClaimID == id
	}
}

func byStatus(st Status) comparator {
	return func(c *Claim) bool {
		return c.Status == st
	}
}

func (l list) removeAll(cmp comparator) {

	for op, v := range l {
		if cmp(v) {
			delete(l, op)
		}
	}
}

func (l list) find(cmp comparator) *Claim {

	for _, v := range l {
		if cmp(v) {
			return v
		}
	}

	return nil
}
