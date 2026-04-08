package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/local/decentralized-sla/aggregator/chain"
)

// startAPI starts HTTP server on cfg.API.Port
func startAPI(db *sql.DB, chainClient *chain.Client, outputDir string, port int, contractAddr string) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name":    "Decentralized SLA Aggregator API",
			"status":  "online",
			"routes": []string{"/", "/health", "/root?bucket_id={id}", "/proof?url={url}&timestamp={ts}", "/chain/root?bucket_id={id}"},
		})
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "healthy"})
	})

	http.HandleFunc("/root", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		b := q.Get("bucket_id")
		if b == "" {
			http.Error(w, "missing bucket_id", http.StatusBadRequest)
			return
		}
		bid, err := strconv.ParseInt(b, 10, 64)
		if err != nil {
			http.Error(w, "invalid bucket_id", http.StatusBadRequest)
			return
		}
		row := db.QueryRow("SELECT root_hash, leaf_count, tx_hash, contract_address, created_at FROM merkle_roots WHERE bucket_id = ?", bid)
		var root string
		var txHash, cAddr sql.NullString
		var leafCount int
		var created int64
		if err := row.Scan(&root, &leafCount, &txHash, &cAddr, &created); err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "bucket not found", http.StatusNotFound)
				return
			}
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"bucket_id":        bid,
			"root_hash":        root,
			"leaf_count":       leafCount,
			"tx_hash":          txHash.String,
			"contract_address": cAddr.String,
			"created_at":       created,
		})
	})

	http.HandleFunc("/chain/root", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		b := q.Get("bucket_id")
		if b == "" || chainClient == nil {
			http.Error(w, "missing bucket_id or chain disabled", http.StatusBadRequest)
			return
		}
		bid, _ := strconv.ParseInt(b, 10, 64)
		root, err := chainClient.GetRoot(bid)
		if err != nil {
			http.Error(w, "chain error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"bucket_id":    bid,
			"onchain_root": "0x" + root,
			"source":       "blockchain",
		})
	})

	http.HandleFunc("/proof", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		url := q.Get("url")
		timestamp := q.Get("timestamp")
		if url == "" || timestamp == "" {
			http.Error(w, "missing url or timestamp", http.StatusBadRequest)
			return
		}
		ts, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			http.Error(w, "invalid timestamp", http.StatusBadRequest)
			return
		}
		bucket := ts / 60
		// Find matching raw_log
		row := db.QueryRow("SELECT id, signature, public_key, leaf_hash, url, status, latency_ms, timestamp, response_hash, watcher_id, error_type FROM raw_logs WHERE url = ? AND timestamp = ? LIMIT 1", url, ts)
		var id int64
		var signature, pubkey, leafHash, rurl string
		var status int
		var latency int64
		var tsdb int64
		var responseHash string
		var watcherID, errorType sql.NullString
		if err := row.Scan(&id, &signature, &pubkey, &leafHash, &rurl, &status, &latency, &tsdb, &responseHash, &watcherID, &errorType); err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "record not found", http.StatusNotFound)
				return
			}
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		// Load proof file for bucket
		fn := filepath.Join(outputDir, fmt.Sprintf("%d.json", bucket))
		bts, err := ioutil.ReadFile(fn)
		if err != nil {
			http.Error(w, "proof file not found", http.StatusInternalServerError)
			return
		}
		var out map[string]interface{}
		if err := json.Unmarshal(bts, &out); err != nil {
			http.Error(w, "invalid proof file", http.StatusInternalServerError)
			return
		}
		proofs, ok := out["proofs"].(map[string]interface{})
		if !ok {
			http.Error(w, "invalid proofs structure", http.StatusInternalServerError)
			return
		}
		pRaw, ok := proofs[fmt.Sprintf("%d", id)]
		if !ok {
			http.Error(w, "proof not found for record", http.StatusNotFound)
			return
		}
		// decode proof object into MerkleProof
		pBytes, err := json.Marshal(pRaw)
		if err != nil {
			http.Error(w, "invalid proof data", http.StatusInternalServerError)
			return
		}
		var mp MerkleProof
		if err := json.Unmarshal(pBytes, &mp); err != nil {
			http.Error(w, "invalid proof format", http.StatusInternalServerError)
			return
		}

		// Verify signature
		signedRec := SignedResult{
			Data: MonitoringResult{
				URL:          rurl,
				Status:       status,
				LatencyMs:    latency,
				Timestamp:    tsdb,
				ResponseHash: responseHash,
				WatcherID:    watcherID.String,
				ErrorType:    errorType.String,
			},
			Signature: signature,
			PublicKey: pubkey,
		}
		sigErr := VerifySignedResult(signedRec)
		sigValid := (sigErr == nil)

		// Verify merkle proof
		merkleValid := mp.Verify()

		// Fetch on-chain root for verification
		var onchainRoot string
		if chainClient != nil {
			onchainRoot, _ = chainClient.GetRoot(bucket)
		}
		onchainMatch := false
		if onchainRoot != "" {
			// normalize mp.Root (could have 0x)
			rootClean := strings.TrimPrefix(mp.Root, "0x")
			onchainMatch = (rootClean == onchainRoot)
		}

		// Return record + proof + verification
		resp := map[string]interface{}{
			"record": map[string]interface{}{
				"data": map[string]interface{}{
					"url":           rurl,
					"status":        status,
					"latency_ms":    latency,
					"timestamp":     tsdb,
					"timestamp_fmt": time.Unix(tsdb, 0).Format(time.RFC3339),
					"response_hash": responseHash,
					"watcher_id":    watcherID.String,
				},
				"signature":  signature,
				"public_key": pubkey,
			},
			"proof":        mp,
			"root":         mp.Root,
			"onchain_root": "0x" + onchainRoot,
			"verification": map[string]interface{}{
				"signature_valid":      sigValid,
				"merkle_proof_valid":   merkleValid,
				"onchain_root_matches": onchainMatch,
			},
			"contract_address": contractAddr,
		}
		json.NewEncoder(w).Encode(resp)
	})

	addr := fmt.Sprintf(":%d", port)
	fmt.Println("Starting API on", addr)
	return http.ListenAndServe(addr, nil)
}
