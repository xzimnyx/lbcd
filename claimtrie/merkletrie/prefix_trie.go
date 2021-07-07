package merkletrie

import (
	"github.com/lbryio/chain/chaincfg/chainhash"
)

type KeyType []byte

type PrefixTrieNode struct { // implements sort.Interface
	children  []*PrefixTrieNode
	key       KeyType
	hash      *chainhash.Hash
	hasClaims bool
}

// insertAt inserts v into s at index i and returns the new slice.
// https://stackoverflow.com/questions/42746972/golang-insert-to-a-sorted-slice
func insertAt(data []*PrefixTrieNode, i int, v *PrefixTrieNode) []*PrefixTrieNode {
	if i == len(data) {
		// Insert at end is the easy case.
		return append(data, v)
	}

	// Make space for the inserted element by shifting
	// values at the insertion index up one index. The call
	// to append does not allocate memory when cap(data) is
	// greater than len(data).
	data = append(data[:i+1], data[i:]...)
	data[i] = v
	return data
}

func (ptn *PrefixTrieNode) Insert(value *PrefixTrieNode) *PrefixTrieNode {
	// keep it sorted (and sort.Sort is too slow)
	index := sortSearch(ptn.children, value.key[0])
	ptn.children = insertAt(ptn.children, index, value)

	return value
}

// this sort.Search is stolen shamelessly from search.go,
// and modified for performance to not need a closure
func sortSearch(nodes []*PrefixTrieNode, b byte) int {
	i, j := 0, len(nodes)
	for i < j {
		h := int(uint(i+j) >> 1) // avoid overflow when computing h
		// i ≤ h < j
		if nodes[h].key[0] < b {
			i = h + 1 // preserves f(i-1) == false
		} else {
			j = h // preserves f(j) == true
		}
	}
	// i == j, f(i-1) == false, and f(j) (= f(i)) == true  =>  answer is i.
	return i
}

func (ptn *PrefixTrieNode) FindNearest(start KeyType) (int, *PrefixTrieNode) {
	// none of the children overlap on the first char or we would have a parent node with that char
	index := sortSearch(ptn.children, start[0])
	hits := ptn.children[index:]
	if len(hits) > 0 {
		return index, hits[0]
	}
	return -1, nil
}

type PrefixTrie interface {
	InsertOrFind(value KeyType) (bool, *PrefixTrieNode)
	Find(value KeyType) *PrefixTrieNode
	FindPath(value KeyType) ([]int, []*PrefixTrieNode)
	IterateFrom(start KeyType, handler func(value *PrefixTrieNode) bool)
	Erase(value KeyType) bool
	NodeCount() int
}

type prefixTrie struct {
	root  *PrefixTrieNode
	Nodes int
}

func NewPrefixTrie() PrefixTrie {
	// we never delete the root node
	return &prefixTrie{root: &PrefixTrieNode{key: make(KeyType, 0)}, Nodes: 1}
}

func (pt *prefixTrie) NodeCount() int {
	return pt.Nodes
}

func matchLength(a, b KeyType) int {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	for i := 0; i < minLen; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return minLen
}

func (pt *prefixTrie) insert(value KeyType, node *PrefixTrieNode) (bool, *PrefixTrieNode) {
	index, child := node.FindNearest(value)
	match := 0
	if index >= 0 { // if we found a child
		match = matchLength(value, child.key)
		if len(value) == match && len(child.key) == match {
			return false, child
		}
	}
	if match <= 0 {
		pt.Nodes++
		return true, node.Insert(&PrefixTrieNode{key: value})
	}
	if match < len(child.key) {
		grandChild := PrefixTrieNode{key: child.key[match:], children: child.children,
			hasClaims: child.hasClaims, hash: child.hash}
		newChild := PrefixTrieNode{key: child.key[0:match], children: []*PrefixTrieNode{&grandChild}}
		child = &newChild
		node.children[index] = child
		pt.Nodes++
		if len(value) == match {
			return true, child
		}
	}
	return pt.insert(value[match:], child)
}

func (pt *prefixTrie) InsertOrFind(value KeyType) (bool, *PrefixTrieNode) {
	if len(value) <= 0 {
		return false, pt.root
	}
	return pt.insert(value, pt.root)
}

func find(value KeyType, node *PrefixTrieNode, pathIndexes *[]int, path *[]*PrefixTrieNode) *PrefixTrieNode {
	index, child := node.FindNearest(value)
	if index < 0 {
		return nil
	}
	match := matchLength(value, child.key)
	if len(value) == match && len(child.key) == match {
		if pathIndexes != nil {
			*pathIndexes = append(*pathIndexes, index)
		}
		if path != nil {
			*path = append(*path, child)
		}
		return child
	}
	if match < len(child.key) || match == len(value) {
		return nil
	}
	if pathIndexes != nil {
		*pathIndexes = append(*pathIndexes, index)
	}
	if path != nil {
		*path = append(*path, child)
	}
	return find(value[match:], child, pathIndexes, path)
}

func (pt *prefixTrie) Find(value KeyType) *PrefixTrieNode {
	if len(value) <= 0 {
		return pt.root
	}
	return find(value, pt.root, nil, nil)
}

func (pt *prefixTrie) FindPath(value KeyType) ([]int, []*PrefixTrieNode) {
	pathIndexes := []int{-1}
	path := []*PrefixTrieNode{pt.root}
	result := find(value, pt.root, &pathIndexes, &path)
	if result == nil {
		return nil, nil
	} // not sure I want this line
	return pathIndexes, path
}

// IterateFrom can be used to find a value and run a function on that value.
// If the handler returns true it continues to iterate through the children of value.
func (pt *prefixTrie) IterateFrom(start KeyType, handler func(value *PrefixTrieNode) bool) {
	node := find(start, pt.root, nil, nil)
	if node == nil {
		return
	}
	iterateFrom(node, handler)
}

func iterateFrom(node *PrefixTrieNode, handler func(value *PrefixTrieNode) bool) {
	for handler(node) {
		for _, child := range node.children {
			iterateFrom(child, handler)
		}
	}
}

func (pt *prefixTrie) Erase(value KeyType) bool {
	indexes, path := pt.FindPath(value)
	if path == nil || len(path) <= 1 {
		return false
	}
	nodes := pt.Nodes
	for i := len(path) - 1; i > 0; i-- {
		childCount := len(path[i].children)
		noClaimData := !path[i].hasClaims
		if childCount == 1 && noClaimData {
			path[i].key = append(path[i].key, path[i].children[0].key...)
			path[i].hash = nil
			path[i].hasClaims = path[i].children[0].hasClaims
			path[i].children = path[i].children[0].children
			pt.Nodes--
			continue
		}
		if childCount == 0 && noClaimData {
			index := indexes[i]
			path[i-1].children = append(path[i-1].children[:index], path[i-1].children[index+1:]...)
			pt.Nodes--
			continue
		}
		break
	}
	return nodes > pt.Nodes
}
