package main

import (
	"os"
	"testing"
)

func TestIntegrationLoadAndBuild(t *testing.T) {
	// create temp DB
	dbPath := "./db/test_integration.db"
	os.Remove(dbPath)
	db, err := InitDB(dbPath)
	if err != nil { t.Fatalf("db init: %v", err) }
	defer func(){ db.Close(); os.Remove(dbPath) }()

	// parse watcher logs (use watcher-node/logs from project root -- tests expect files present)
	n, err := parseWatcherLogs(db, "../watcher-node/logs")
	if err != nil { t.Fatalf("parse logs: %v", err) }
	if n == 0 { t.Fatalf("expected at least 1 record from logs") }

	// find buckets and build
	rows, err := db.Query("SELECT DISTINCT bucket_id FROM raw_logs")
	if err != nil { t.Fatalf("select buckets: %v", err) }
	defer rows.Close()
	for rows.Next() {
		var b int64
		rows.Scan(&b)
		if err := BuildMerkleForBucket(db, b, "./output", nil, ""); err != nil {
			t.Fatalf("build bucket %d: %v", b, err)
		}
	}
}
