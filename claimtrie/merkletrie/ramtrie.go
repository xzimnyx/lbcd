package merkletrie

import (
	"bytes"
	"errors"
	"runtime"
	"sync"

	"github.com/lbryio/lbcd/chaincfg/chainhash"
	"github.com/lbryio/lbcd/claimtrie/node"
)

type MerkleTrie interface {
	SetRoot(h *chainhash.Hash) error
	Update(name []byte, h *chainhash.Hash, restoreChildren bool)
	MerkleHash() *chainhash.Hash
	MerkleHashAllClaims() *chainhash.Hash
	Flush() error
}

type RamTrie struct {
	collapsedTrie
	bufs *sync.Pool
}

func NewRamTrie() *RamTrie {
	return &RamTrie{
		bufs: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
		collapsedTrie: collapsedTrie{Root: &collapsedVertex{}},
	}
}

type ramTriePayload struct {
	merkleHash *chainhash.Hash
	claimHash  *chainhash.Hash
}

func (r *ramTriePayload) clear() {
	r.claimHash = nil
	r.merkleHash = nil
}

func (r *ramTriePayload) childModified() {
	r.merkleHash = nil
}

func (r *ramTriePayload) isEmpty() bool {
	return r.claimHash == nil
}

func getOrMakePayload(v *collapsedVertex) *ramTriePayload {
	if v.payload == nil {
		r := &ramTriePayload{}
		v.payload = r
		return r
	}
	return v.payload.(*ramTriePayload)
}

var _ VertexPayload = &ramTriePayload{}

var ErrFullRebuildRequired = errors.New("a full rebuild is required")

func (rt *RamTrie) SetRoot(h *chainhash.Hash) error {
	if getOrMakePayload(rt.Root).merkleHash.IsEqual(h) {
		runtime.GC()
		return nil
	}

	// should technically clear the old trie first, but this is abused for partial rebuilds so don't
	return ErrFullRebuildRequired
}

func (rt *RamTrie) Update(name []byte, h *chainhash.Hash, _ bool) {
	if h == nil {
		rt.Erase(name)
	} else {
		_, n := rt.InsertOrFind(name)
		getOrMakePayload(n).claimHash = h
	}
}

func (rt *RamTrie) MerkleHash() *chainhash.Hash {
	if h := rt.merkleHash(rt.Root); h == nil {
		return EmptyTrieHash
	}
	return getOrMakePayload(rt.Root).merkleHash
}

func (rt *RamTrie) merkleHash(v *collapsedVertex) *chainhash.Hash {
	p := getOrMakePayload(v)
	if p.merkleHash != nil {
		return p.merkleHash
	}

	b := rt.bufs.Get().(*bytes.Buffer)
	defer rt.bufs.Put(b)
	b.Reset()

	for _, ch := range v.children {
		h := rt.merkleHash(ch)              // h is a pointer; don't destroy its data
		b.WriteByte(ch.key[0])              // nolint : errchk
		b.Write(rt.completeHash(h, ch.key)) // nolint : errchk
	}

	if p.claimHash != nil {
		b.Write(p.claimHash[:])
	}

	if b.Len() > 0 {
		h := chainhash.DoubleHashH(b.Bytes())
		p.merkleHash = &h
	}

	return p.merkleHash
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
	if h := rt.merkleHashAllClaims(rt.Root); h == nil {
		return EmptyTrieHash
	}
	return getOrMakePayload(rt.Root).merkleHash
}

func (rt *RamTrie) merkleHashAllClaims(v *collapsedVertex) *chainhash.Hash {
	p := getOrMakePayload(v)
	if p.merkleHash != nil {
		return p.merkleHash
	}

	childHashes := make([]*chainhash.Hash, 0, len(v.children))
	for _, ch := range v.children {
		h := rt.merkleHashAllClaims(ch)
		childHashes = append(childHashes, h)
	}

	claimHash := NoClaimsHash
	if p.claimHash != nil {
		claimHash = p.claimHash
	} else if len(childHashes) == 0 {
		return nil
	}

	childHash := NoChildrenHash
	if len(childHashes) > 0 {
		// this shouldn't be referencing node; where else can we put this merkle root func?
		childHash = node.ComputeMerkleRoot(childHashes)
	}

	p.merkleHash = node.HashMerkleBranches(childHash, claimHash)
	return p.merkleHash
}

func (rt *RamTrie) Flush() error {
	return nil
}
