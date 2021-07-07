package merkletrie

import (
	"bytes"
	"fmt"
	"sort"
	"sync"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/cockroachdb/pebble"
)

var (
	// EmptyTrieHash represents the Merkle Hash of an empty MerkleTrie.
	// "0000000000000000000000000000000000000000000000000000000000000001"
	EmptyTrieHash  = &chainhash.Hash{1}
	NoChildrenHash = &chainhash.Hash{2}
	NoClaimsHash   = &chainhash.Hash{3}
)

// ValueStore enables MerkleTrie to query node values from different implementations.
type ValueStore interface {
	ClaimHashes(name []byte) []*chainhash.Hash
	Hash(name []byte) *chainhash.Hash
}

// MerkleTrie implements a 256-way prefix tree.
type MerkleTrie struct {
	store ValueStore
	repo  Repo

	root *vertex
	bufs *sync.Pool
}

// New returns a MerkleTrie.
func New(store ValueStore, repo Repo) *MerkleTrie {

	tr := &MerkleTrie{
		store: store,
		repo:  repo,
		bufs: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
		root: newVertex(EmptyTrieHash),
	}

	return tr
}

// SetRoot drops all resolved nodes in the MerkleTrie, and set the root with specified hash.
func (t *MerkleTrie) SetRoot(h *chainhash.Hash) {
	t.root = newVertex(h)
}

// Update updates the nodes along the path to the key.
// Each node is resolved or created with their Hash cleared.
func (t *MerkleTrie) Update(name []byte, restoreChildren bool) {

	n := t.root
	for i, ch := range name {
		if restoreChildren && len(n.childLinks) == 0 {
			t.resolveChildLinks(n, name[:i])
		}
		if n.childLinks[ch] == nil {
			n.childLinks[ch] = newVertex(nil)
		}
		n.merkleHash = nil
		n = n.childLinks[ch]
	}

	if restoreChildren && len(n.childLinks) == 0 {
		t.resolveChildLinks(n, name)
	}
	n.hasValue = true
	n.merkleHash = nil
	n.claimsHash = nil
}

// resolveChildLinks updates the links on n
func (t *MerkleTrie) resolveChildLinks(n *vertex, key []byte) {

	if n.merkleHash == nil {
		return
	}

	b := t.bufs.Get().(*bytes.Buffer)
	defer t.bufs.Put(b)
	b.Reset()
	b.Write(key)
	b.Write(n.merkleHash[:])

	result, closer, err := t.repo.Get(b.Bytes())
	if err == pebble.ErrNotFound { // TODO: leaky abstraction
		return
	} else if err != nil {
		panic(err)
	}
	defer closer.Close()

	nb := nbuf(result)
	n.hasValue, n.claimsHash = nb.hasValue()
	for i := 0; i < nb.entries(); i++ {
		p, h := nb.entry(i)
		n.childLinks[p] = newVertex(h)
	}
}

// MerkleHash returns the Merkle Hash of the MerkleTrie.
// All nodes must have been resolved before calling this function.
func (t *MerkleTrie) MerkleHash() *chainhash.Hash {
	buf := make([]byte, 0, 256)
	if h := t.merkle(buf, t.root); h == nil {
		return EmptyTrieHash
	}
	return t.root.merkleHash
}

// merkle recursively resolves the hashes of the node.
// All nodes must have been resolved before calling this function.
func (t *MerkleTrie) merkle(prefix []byte, v *vertex) *chainhash.Hash {
	if v.merkleHash != nil {
		return v.merkleHash
	}

	b := t.bufs.Get().(*bytes.Buffer)
	defer t.bufs.Put(b)
	b.Reset()

	keys := keysInOrder(v)

	for _, ch := range keys {
		child := v.childLinks[ch]
		if child == nil {
			continue
		}
		p := append(prefix, ch)
		h := t.merkle(p, child)
		if h != nil {
			b.WriteByte(ch) // nolint : errchk
			b.Write(h[:])   // nolint : errchk
		}
		if h == nil || len(prefix) > 4 { // TODO: determine the right number here
			delete(v.childLinks, ch) // keep the RAM down (they get recreated on Update)
		}
	}

	if v.hasValue {
		claimHash := v.claimsHash
		if claimHash == nil {
			claimHash = t.store.Hash(prefix)
			v.claimsHash = claimHash
		}
		if claimHash != nil {
			b.Write(claimHash[:])
		} else {
			v.hasValue = false
		}
	}

	if b.Len() > 0 {
		h := chainhash.DoubleHashH(b.Bytes())
		v.merkleHash = &h
		t.repo.Set(append(prefix, h[:]...), b.Bytes())
	}

	return v.merkleHash
}

func keysInOrder(v *vertex) []byte {
	keys := make([]byte, 0, len(v.childLinks))
	for key := range v.childLinks {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func (t *MerkleTrie) MerkleHashAllClaims() *chainhash.Hash {
	buf := make([]byte, 0, 256)
	if h := t.merkleAllClaims(buf, t.root); h == nil {
		return EmptyTrieHash
	}
	return t.root.merkleHash
}

func (t *MerkleTrie) merkleAllClaims(prefix []byte, v *vertex) *chainhash.Hash {
	if v.merkleHash != nil {
		return v.merkleHash
	}
	b := t.bufs.Get().(*bytes.Buffer)
	defer t.bufs.Put(b)
	b.Reset()

	keys := keysInOrder(v)
	childHashes := make([]*chainhash.Hash, 0, len(keys))
	for _, ch := range keys {
		n := v.childLinks[ch]
		if n == nil {
			continue
		}
		p := append(prefix, ch)
		h := t.merkleAllClaims(p, n)
		if h != nil {
			childHashes = append(childHashes, h)
			b.WriteByte(ch) // nolint : errchk
			b.Write(h[:])   // nolint : errchk
		}
		if h == nil || len(prefix) > 4 { // TODO: determine the right number here
			delete(v.childLinks, ch) // keep the RAM down (they get recreated on Update)
		}
	}

	var claimsHash *chainhash.Hash
	if v.hasValue {
		claimsHash = v.claimsHash
		if claimsHash == nil {
			claimHashes := t.store.ClaimHashes(prefix)
			if len(claimHashes) > 0 {
				claimsHash = computeMerkleRoot(claimHashes)
				v.claimsHash = claimsHash
			} else {
				v.hasValue = false
			}
		}
	}

	if len(childHashes) > 1 || claimsHash != nil { // yeah, about that 1 there -- old code used the condensed trie
		left := NoChildrenHash
		if len(childHashes) > 0 {
			left = computeMerkleRoot(childHashes)
		}
		right := NoClaimsHash
		if claimsHash != nil {
			b.Write(claimsHash[:]) // for Has Value, nolint : errchk
			right = claimsHash
		}

		h := hashMerkleBranches(left, right)
		v.merkleHash = h
		t.repo.Set(append(prefix, h[:]...), b.Bytes())
	} else if len(childHashes) == 1 {
		v.merkleHash = childHashes[0] // pass it up the tree
		t.repo.Set(append(prefix, v.merkleHash[:]...), b.Bytes())
	}

	return v.merkleHash
}

func (t *MerkleTrie) Close() error {
	return t.repo.Close()
}

func (t *MerkleTrie) Dump(s string, allClaims bool) {
	v := t.root

	for i := 0; i < len(s); i++ {
		t.resolveChildLinks(v, []byte(s[:i]))
		ch := s[i]
		v = v.childLinks[ch]
		if v == nil {
			fmt.Printf("Missing child at %s\n", s[:i+1])
			return
		}
	}
	t.resolveChildLinks(v, []byte(s))

	fmt.Printf("Node hash: %s, has value: %t\n", v.merkleHash.String(), v.hasValue)

	for key, value := range v.childLinks {
		fmt.Printf("  Child %s hash: %s\n", string(key), value.merkleHash.String())
	}
}
