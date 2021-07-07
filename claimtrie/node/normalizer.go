package node

import (
	"github.com/btcsuite/btcd/claimtrie/param"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

//func init() {
//	if cases.UnicodeVersion[:2] != "11" {
//		panic("Wrong unicode version!")
//	}
//}

var Normalize = normalizeGo

func NormalizeIfNecessary(name []byte, height int32) []byte {
	if height < param.ActiveParams.NormalizedNameForkHeight {
		return name
	}
	return Normalize(name)
}

var folder = cases.Fold()

func normalizeGo(value []byte) []byte {

	normalized := norm.NFD.Bytes(value)
	return folder.Bytes(normalized)
}
