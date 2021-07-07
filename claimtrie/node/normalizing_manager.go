package node

import (
	"bytes"
	"fmt"

	"github.com/btcsuite/btcd/claimtrie/change"
	"github.com/btcsuite/btcd/claimtrie/param"
)

type NormalizingManager struct { // implements Manager
	Manager
	normalizedAt int32
}

func NewNormalizingManager(baseManager Manager) Manager {
	return &NormalizingManager{
		Manager:      baseManager,
		normalizedAt: -1,
	}
}

func (nm *NormalizingManager) AppendChange(chg change.Change) error {
	chg.Name = NormalizeIfNecessary(chg.Name, chg.Height)
	return nm.Manager.AppendChange(chg)
}

func (nm *NormalizingManager) IncrementHeightTo(height int32) ([][]byte, error) {
	nm.addNormalizationForkChangesIfNecessary(height)
	return nm.Manager.IncrementHeightTo(height)
}

func (nm *NormalizingManager) DecrementHeightTo(affectedNames [][]byte, height int32) error {
	if nm.normalizedAt > height {
		nm.normalizedAt = -1
	}
	return nm.Manager.DecrementHeightTo(affectedNames, height)
}

func (nm *NormalizingManager) NextUpdateHeightOfNode(name []byte) ([]byte, int32) {
	name, nextUpdate := nm.Manager.NextUpdateHeightOfNode(name)
	if nextUpdate > param.NormalizedNameForkHeight {
		name = Normalize(name)
	}
	return name, nextUpdate
}

func (nm *NormalizingManager) addNormalizationForkChangesIfNecessary(height int32) {

	if nm.Manager.Height()+1 != height {
		// initialization phase
		if height >= param.NormalizedNameForkHeight {
			nm.normalizedAt = param.NormalizedNameForkHeight // eh, we don't really know that it happened there
		}
	}

	if nm.normalizedAt >= 0 || height != param.NormalizedNameForkHeight {
		return
	}
	nm.normalizedAt = height
	fmt.Printf("Generating necessary changes for the normalization fork...\n")

	// the original code had an unfortunate bug where many unnecessary takeovers
	// were triggered at the normalization fork
	predicate := func(name []byte) bool {
		norm := Normalize(name)
		eq := bytes.Equal(name, norm)
		if eq {
			return true
		}

		clone := make([]byte, len(name))
		copy(clone, name) // iteration name buffer is reused on future loops

		// by loading changes for norm here, you can determine if there will be a conflict

		n, err := nm.Manager.Node(clone)
		if err != nil || n == nil {
			return true
		}
		for _, c := range n.Claims {
			nm.Manager.AppendChange(change.Change{
				Type:          change.AddClaim,
				Name:          norm,
				Height:        c.AcceptedAt,
				OutPoint:      c.OutPoint.String(),
				ClaimID:       c.ClaimID,
				Amount:        c.Amount,
				Value:         c.Value,
				ActiveHeight:  c.ActiveAt, // necessary to match the old hash
				VisibleHeight: height,     // necessary to match the old hash; it would have been much better without
			})
			nm.Manager.AppendChange(change.Change{
				Type:     change.SpendClaim,
				Name:     clone,
				Height:   height,
				OutPoint: c.OutPoint.String(),
			})
		}
		for _, c := range n.Supports {
			nm.Manager.AppendChange(change.Change{
				Type:          change.AddSupport,
				Name:          norm,
				Height:        c.AcceptedAt,
				OutPoint:      c.OutPoint.String(),
				ClaimID:       c.ClaimID,
				Amount:        c.Amount,
				Value:         c.Value,
				ActiveHeight:  c.ActiveAt,
				VisibleHeight: height,
			})
			nm.Manager.AppendChange(change.Change{
				Type:     change.SpendSupport,
				Name:     clone,
				Height:   height,
				OutPoint: c.OutPoint.String(),
			})
		}

		return true
	}
	nm.Manager.IterateNames(predicate)
}
