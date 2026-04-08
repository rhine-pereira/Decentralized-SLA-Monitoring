package main

import "testing"

func TestMerkleEvenLeaves(t *testing.T) {
	leaves := []string{
		"a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1",
		"b1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f1b2",
		"c1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f2c3",
		"d1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f3d4",
	}
	mt, err := NewMerkleTree(leaves)
	if err != nil { t.Fatalf("failed build: %v", err) }
	if mt.Root() == "" { t.Fatalf("expected non-empty root") }
	p, err := mt.GetProof(2)
	if err != nil { t.Fatalf("getproof: %v", err) }
	if !p.Verify() { t.Fatalf("proof verify failed") }
}

func TestMerkleOddLeaves(t *testing.T) {
	leaves := []string{
		"a1",
		"b1",
		"c1",
	}
	mt, err := NewMerkleTree(leaves)
	if err != nil { t.Fatalf("failed build: %v", err) }
	p, err := mt.GetProof(2)
	if err != nil { t.Fatalf("getproof: %v", err) }
	if !p.Verify() { t.Fatalf("proof verify failed") }
}

func TestProofSingleLeaf(t *testing.T) {
	leaves := []string{"a1"}
	mt, _ := NewMerkleTree(leaves)
	p, _ := mt.GetProof(0)
	if !p.Verify() { t.Fatalf("single leaf proof failed") }
}
