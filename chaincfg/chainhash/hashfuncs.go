// Copyright (c) 2015 The Decred developers
// Copyright (c) 2016-2017 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chainhash

import (
	"crypto/sha256"
	"crypto/sha512"
	"golang.org/x/crypto/ripemd160"
)

// HashB calculates hash(b) and returns the resulting bytes.
func HashB(b []byte) []byte {
	hash := sha256.Sum256(b)
	return hash[:]
}

// HashH calculates hash(b) and returns the resulting bytes as a Hash.
func HashH(b []byte) Hash {
	return Hash(sha256.Sum256(b))
}

// DoubleHashB calculates hash(hash(b)) and returns the resulting bytes.
func DoubleHashB(b []byte) []byte {
	first := sha256.Sum256(b)
	second := sha256.Sum256(first[:])
	return second[:]
}

// DoubleHashH calculates hash(hash(b)) and returns the resulting bytes as a
// Hash.
func DoubleHashH(b []byte) Hash {
	first := sha256.Sum256(b)
	return Hash(sha256.Sum256(first[:]))
}

// LbryPoWHashH calculates returns the PoW Hash.
//
//   doubled  := SHA256(SHA256(b))
//   expanded := SHA512(doubled)
//   left     := RIPEMD160(expanded[0:32])
//   right    := RIPEMD160(expanded[32:64])
//   result   := SHA256(SHA256(left||right))
func LbryPoWHashH(b []byte) Hash {
	doubled := DoubleHashB(b)
	expanded := sha512.Sum512(doubled)

	r := ripemd160.New()
	r.Reset()
	r.Write(expanded[:sha256.Size])
	left := r.Sum(nil)

	r.Reset()
	r.Write(expanded[sha256.Size:])

	combined := r.Sum(left)
	return DoubleHashH(combined)
}

func HashMerkleBranches(left *Hash, right *Hash) *Hash {
	// Concatenate the left and right nodes.
	var hash [HashSize * 2]byte
	copy(hash[:HashSize], left[:])
	copy(hash[HashSize:], right[:])

	newHash := DoubleHashH(hash[:])
	return &newHash
}

func ComputeMerkleRoot(hashes []*Hash) *Hash {
	if len(hashes) <= 0 {
		return nil
	}
	for len(hashes) > 1 {
		if (len(hashes) & 1) > 0 { // odd count
			hashes = append(hashes, hashes[len(hashes)-1])
		}
		for i := 0; i < len(hashes); i += 2 { // TODO: parallelize this loop (or use a lib that does it)
			hashes[i>>1] = HashMerkleBranches(hashes[i], hashes[i+1])
		}
		hashes = hashes[:len(hashes)>>1]
	}
	return hashes[0]
}
