package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/local/decentralized-sla/aggregator/chain"
)

// BuildMerkleForBucket reads all leaf_hashes for a bucket, constructs a Merkle tree, generates proofs for each leaf, writes output JSON file, and stores root in DB.
func BuildMerkleForBucket(db *sql.DB, bucket int64, outputDir string, chainClient *chain.Client, contractAddr string) error {
	rows, err := db.Query("SELECT id, leaf_hash FROM raw_logs WHERE bucket_id = ? ORDER BY id ASC", bucket)
	if err != nil {
		return err
	}
	defer rows.Close()

	leafHashes := []string{}
	ids := []int64{}
	for rows.Next() {
		var id int64
		var leaf string
		if err := rows.Scan(&id, &leaf); err != nil {
			return err
		}
		ids = append(ids, id)
		leafHashes = append(leafHashes, leaf)
	}
	if len(leafHashes) == 0 {
		fmt.Println("no leaves for bucket", bucket)
		return nil
	}

	mt, err := NewMerkleTree(leafHashes)
	if err != nil {
		return err
	}
	root := mt.Root()

	// generate proofs
	proofs := map[string]MerkleProof{}
	for i := range leafHashes {
		pr, err := mt.GetProof(i)
		if err != nil {
			return err
		}
		key := fmt.Sprintf("%d", ids[i])
		proofs[key] = pr
	}

	// prepare output structure
	out := map[string]interface{}{
		"bucket_id": bucket,
		"root":      root,
		"leaves":    leafHashes,
		"proofs":    proofs,
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	fn := filepath.Join(outputDir, fmt.Sprintf("%d.json", bucket))
	if err := ioutil.WriteFile(fn, b, 0644); err != nil {
		return err
	}

	// insert or replace root
	_, err = db.Exec(`INSERT OR REPLACE INTO merkle_roots (bucket_id, root_hash, leaf_count) VALUES (?,?,?)`, bucket, root, len(leafHashes))
	if err != nil {
		return err
	}

	fmt.Println("Wrote merkle output for bucket", bucket, "->", fn)

	// Anchoring on-chain
	if chainClient != nil {
		txHash, err := chainClient.StoreRoot(bucket, root)
		if err != nil {
			fmt.Printf("warning: chain store failed for bucket %d: %v\n", bucket, err)
		} else if txHash != "" {
			if txHash == "ALREADY_STORED" {
				fmt.Printf("Bucket %d already anchored on-chain\n", bucket)
				db.Exec(`UPDATE merkle_roots SET contract_address = ? WHERE bucket_id = ? AND (contract_address IS NULL OR contract_address = '')`, contractAddr, bucket)
			} else {
				fmt.Printf("Anchored bucket %d on-chain, tx: %s\n", bucket, txHash)
				db.Exec(`UPDATE merkle_roots SET tx_hash = ?, contract_address = ? WHERE bucket_id = ?`, txHash, contractAddr, bucket)
			}
		}
	}

	return nil
}
