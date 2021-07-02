package merkletrie

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"

	"github.com/stretchr/testify/require"
)

func TestName(t *testing.T) {

	r := require.New(t)

	target, _ := chainhash.NewHashFromStr("e9ffb584c62449f157c8be88257bd1eebb2d8ef824f5c86b43c4f8fd9e800d6a")

	data := []*chainhash.Hash{EmptyTrieHash}
	root := chainhash.ComputeMerkleRoot(data)
	r.True(EmptyTrieHash.IsEqual(root))

	data = append(data, NoChildrenHash, NoClaimsHash)
	root = chainhash.ComputeMerkleRoot(data)
	r.True(target.IsEqual(root))
}
