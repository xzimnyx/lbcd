package merkletrie

import (
	"bytes"
	"runtime"
	"strconv"
	"sync"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/claimtrie/node"
)

type MerkleTrie interface {
	SetRoot(h *chainhash.Hash, names [][]byte)
	Update(name []byte, restoreChildren bool)
	MerkleHash() *chainhash.Hash
	MerkleHashAllClaims() *chainhash.Hash
	Flush() error
}

type RamTrie struct {
	collapsedTrie
	store ValueStore
	bufs  *sync.Pool
}

func NewRamTrie(s ValueStore) *RamTrie {
	return &RamTrie{
		store: s,
		bufs: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
		collapsedTrie: collapsedTrie{Root: &collapsedVertex{}},
	}
}

func (rt *RamTrie) SetRoot(h *chainhash.Hash, names [][]byte) {
	if rt.Root.merkleHash.IsEqual(h) {
		runtime.GC()
		return
	}

	// if names is nil then we need to query all names
	if names == nil {
		node.LogOnce("Building the entire claim trie in RAM...") // could put this in claimtrie.go

		//should technically clear the old trie first:
		if rt.Nodes > 1 {
			rt.Root = &collapsedVertex{key: make(KeyType, 0)}
			rt.Nodes = 1
			runtime.GC()
		}

		c := 0
		rt.store.IterateNames(func(name []byte) bool {
			rt.Update(name, false)
			c++
			return true
		})

		node.LogOnce("Completed claim trie construction. Name count: " + strconv.Itoa(c))
	} else {
		for _, name := range names {
			rt.Update(name, false)
		}
	}
	runtime.GC()
}

func (rt *RamTrie) Update(name []byte, _ bool) {
	h := rt.store.Hash(name)
	if h == nil {
		rt.Erase(name)
	} else {
		_, n := rt.InsertOrFind(name)
		n.claimHash = h
	}
}

func (rt *RamTrie) MerkleHash() *chainhash.Hash {
	if h := rt.merkleHash(rt.Root); h == nil {
		return EmptyTrieHash
	}
	return rt.Root.merkleHash
}

func (rt *RamTrie) merkleHash(v *collapsedVertex) *chainhash.Hash {
	if v.merkleHash != nil {
		return v.merkleHash
	}

	b := rt.bufs.Get().(*bytes.Buffer)
	defer rt.bufs.Put(b)
	b.Reset()

	for _, ch := range v.children {
		h := rt.merkleHash(ch)              // h is a pointer; don't destroy its data
		b.WriteByte(ch.key[0])              // nolint : errchk
		b.Write(rt.completeHash(h, ch.key)) // nolint : errchk
	}

	if v.claimHash != nil {
		b.Write(v.claimHash[:])
	}

	if b.Len() > 0 {
		h := chainhash.DoubleHashH(b.Bytes())
		v.merkleHash = &h
	}

	return v.merkleHash
}

func (rt *RamTrie) completeHash(h *chainhash.Hash, childKey KeyType) []byte {
	var data [chainhash.HashSize + 1]byte
	copy(data[1:], h[:])
	for i := len(childKey) - 1; i > 0; i-- {
		data[0] = childKey[i]
		copy(data[1:], chainhash.DoubleHashB(data[:]))
	}
	return data[1:]
}

func (rt *RamTrie) MerkleHashAllClaims() *chainhash.Hash {
	return rt.merkleHashAllClaims(rt.Root)
}

func (rt *RamTrie) merkleHashAllClaims(v *collapsedVertex) *chainhash.Hash {
	if v.merkleHash != nil {
		return v.merkleHash
	}

	childHashes := make([]*chainhash.Hash, 0, len(v.children))
	for _, ch := range v.children {
		h := rt.merkleHashAllClaims(ch)
		childHashes = append(childHashes, h)
	}

	claimHash := NoClaimsHash
	if v.claimHash != nil {
		claimHash = v.claimHash
	} else if len(childHashes) == 0 {
		return v.merkleHash
	}

	childHash := NoChildrenHash
	if len(childHashes) > 0 {
		// this shouldn't be referencing node; where else can we put this merkle root func?
		childHash = node.ComputeMerkleRoot(childHashes)
	}

	v.merkleHash = node.HashMerkleBranches(childHash, claimHash)
	return v.merkleHash
}

func (rt *RamTrie) Flush() error {
	return nil
}
