package main

import (
	"crypto/sha256"
	"encoding/hex"
)

// MerkleProof represents a path from a leaf to the root.
type MerkleProof struct {
	LeafIndex int      `json:"leaf_index"`
	LeafHash  string   `json:"leaf_hash"`
	Siblings  []string `json:"siblings"` // sibling hashes from leaf->root
	Root      string   `json:"root"`
}

// Verify recomputes the path from the leaf using siblings and returns true if final hash equals Root.
func (p *MerkleProof) Verify() bool {
	cur, err := hex.DecodeString(p.LeafHash)
	if err != nil {
		return false
	}
	index := p.LeafIndex
	for _, sibHex := range p.Siblings {
		sib, err := hex.DecodeString(sibHex)
		if err != nil {
			return false
		}
		var buf []byte
		if index%2 == 0 {
			// current is left
			buf = append(buf, cur...)
			buf = append(buf, sib...)
		} else {
			// current is right
			buf = append(buf, sib...)
			buf = append(buf, cur...)
		}
		h := sha256.Sum256(buf)
		cur = h[:]
		index = index / 2
	}
	finalHex := hex.EncodeToString(cur)
	return finalHex == p.Root
}
