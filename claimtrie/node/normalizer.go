package node

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"

	"github.com/btcsuite/btcd/claimtrie/param"
)

var Normalize = normalizeGo

func NormalizeIfNecessary(name []byte, height int32) []byte {
	if height < param.NormalizedNameForkHeight {
		return name
	}
	return Normalize(name)
}

func normalizeGo(value []byte) []byte {

	normalized := norm.NFD.Bytes(value)
	return cases.Fold().Bytes(normalized)
}
