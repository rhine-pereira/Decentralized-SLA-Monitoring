package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

// MerkleTree holds layers of the tree. Layers[0] == leaves (hex hashes), last layer contains root.
type MerkleTree struct {
	Leaves []string   `json:"leaves"`
	Layers [][]string `json:"layers"`
}

// hashPair decodes two hex hashes, concatenates their bytes (left||right) and returns SHA256 hex string.
func hashPair(leftHex, rightHex string) (string, error) {
	left, err := hex.DecodeString(leftHex)
	if err != nil {
		return "", err
	}
	right, err := hex.DecodeString(rightHex)
	if err != nil {
		return "", err
	}
	buf := make([]byte, 0, len(left)+len(right))
	buf = append(buf, left...)
	buf = append(buf, right...)
	h := sha256.Sum256(buf)
	return hex.EncodeToString(h[:]), nil
}

// NewMerkleTree builds a Merkle tree given a slice of leaf hashes (hex-encoded). It returns the tree with layers populated.
func NewMerkleTree(leaves []string) (*MerkleTree, error) {
	if len(leaves) == 0 {
		return &MerkleTree{Leaves: []string{}, Layers: [][]string{}}, nil
	}

	// Work on a copy of leaves so caller's slice isn't mutated
	current := make([]string, len(leaves))
	copy(current, leaves)

	layers := [][]string{}
	layers = append(layers, current)

	for len(current) > 1 {
		if len(current)%2 == 1 {
			// duplicate last element
			current = append(current, current[len(current)-1])
		}
		next := make([]string, 0, len(current)/2)
		for i := 0; i < len(current); i += 2 {
			h, err := hashPair(current[i], current[i+1])
			if err != nil {
				return nil, err
			}
			next = append(next, h)
		}
		layers = append(layers, next)
		current = next
	}

	return &MerkleTree{Leaves: leaves, Layers: layers}, nil
}

// Root returns the Merkle root hex string. If tree has no leaves, returns empty string.
func (t *MerkleTree) Root() string {
	if t == nil || len(t.Layers) == 0 {
		return ""
	}
	last := t.Layers[len(t.Layers)-1]
	if len(last) == 0 {
		return ""
	}
	return last[0]
}

// GetProof generates an inclusion proof for the leaf at index. Returns error if index out of bounds.
func (t *MerkleTree) GetProof(index int) (MerkleProof, error) {
	if t == nil {
		return MerkleProof{}, errors.New("nil merkle tree")
	}
	if index < 0 || index >= len(t.Leaves) {
		return MerkleProof{}, errors.New("index out of bounds")
	}
	siblings := []string{}
	currentIndex := index
	// layers[0] are leaves; ascend until the last layer (root)
	for layer := 0; layer < len(t.Layers)-1; layer++ {
		layerNodes := t.Layers[layer]
		var sibling string
		if currentIndex%2 == 0 {
			// sibling is right
			sibIdx := currentIndex + 1
			if sibIdx >= len(layerNodes) {
				// duplicated last node case
				sibling = layerNodes[currentIndex]
			} else {
				sibling = layerNodes[sibIdx]
			}
		} else {
			// sibling is left
			sibIdx := currentIndex - 1
			if sibIdx < 0 {
				// shouldn't happen, but guard
				sibling = layerNodes[currentIndex]
			} else {
				sibling = layerNodes[sibIdx]
			}
		}
		siblings = append(siblings, sibling)
		currentIndex = currentIndex / 2
	}

	proof := MerkleProof{
		LeafIndex: index,
		LeafHash:  t.Leaves[index],
		Siblings:  siblings,
		Root:      t.Root(),
	}
	return proof, nil
}
