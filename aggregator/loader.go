package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// parseWatcherLogs reads JSONL files from watcher_logs_dir matching pattern watcher_*.jsonl and inserts records into DB.
func parseWatcherLogs(db *sql.DB, logsDir string) (int, error) {
	count := 0
	files, err := filepath.Glob(filepath.Join(logsDir, "watcher_*.jsonl"))
	if err != nil {
		return 0, err
	}
	for _, f := range files {
		fmt.Println("Processing", f)
		fd, err := os.Open(f)
		if err != nil {
			fmt.Println("failed open", err)
			continue
		}
		s := bufio.NewScanner(fd)
		for s.Scan() {
			line := strings.TrimSpace(s.Text())
			if line == "" {
				continue
			}
			var rec SignedResult
			if err := json.Unmarshal([]byte(line), &rec); err != nil {
				fmt.Println("invalid json line", err)
				continue
			}
			leafHash, err := ComputeLeafHash(rec)
			if err != nil {
				fmt.Println("compute leaf hash err", err)
				continue
			}
			bucket := rec.Data.Timestamp / 60
			// insert
			res, err := db.Exec(`INSERT OR IGNORE INTO raw_logs (url,status,latency_ms,timestamp,response_hash,watcher_id,error_type,signature,public_key,bucket_id,leaf_hash) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
				rec.Data.URL, rec.Data.Status, rec.Data.LatencyMs, rec.Data.Timestamp, rec.Data.ResponseHash, rec.Data.WatcherID, rec.Data.ErrorType, rec.Signature, rec.PublicKey, bucket, leafHash)
			if err != nil {
				fmt.Println("insert err", err)
				continue
			}
			affected, _ := res.RowsAffected()
			if affected > 0 {
				count++
			}
		}
		fd.Close()
	}
	return count, nil
}
