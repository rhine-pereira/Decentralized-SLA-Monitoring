package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

// InitDB opens SQLite database (creates file if missing) and runs migrations.
func InitDB(path string) (*sql.DB, error) {
	if err := os.MkdirAll("./db", 0755); err != nil {
		return nil, err
	}
	if !strings.Contains(path, "?") {
		path += "?_pragma=busy_timeout=5000"
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	// Enable WAL mode
	if _, err := db.Exec(`PRAGMA journal_mode=WAL; PRAGMA synchronous=NORMAL;`); err != nil {
		fmt.Printf("warning: failed to set WAL: %v\n", err)
	}
	// Run migrations
	schema := `
CREATE TABLE IF NOT EXISTS raw_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url TEXT NOT NULL,
    status INTEGER,
    latency_ms INTEGER,
    timestamp INTEGER NOT NULL,
    response_hash TEXT,
    watcher_id TEXT NOT NULL,
    error_type TEXT,
    signature TEXT NOT NULL,
    public_key TEXT NOT NULL,
    bucket_id INTEGER NOT NULL,
    leaf_hash TEXT NOT NULL,
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    UNIQUE(url, timestamp, watcher_id)
);

CREATE INDEX IF NOT EXISTS idx_bucket_id ON raw_logs(bucket_id);
CREATE INDEX IF NOT EXISTS idx_timestamp ON raw_logs(timestamp);

CREATE TABLE IF NOT EXISTS merkle_roots (
    bucket_id INTEGER PRIMARY KEY,
    root_hash TEXT NOT NULL,
    leaf_count INTEGER NOT NULL,
    tx_hash TEXT,
    contract_address TEXT,
    created_at INTEGER DEFAULT (strftime('%s', 'now'))
);
`
	_, err = db.Exec(schema)
	if err != nil {
		db.Close()
		return nil, err
	}

	// Add columns to existing DB if they don't exist
	db.Exec(`ALTER TABLE merkle_roots ADD COLUMN tx_hash TEXT;`)
	db.Exec(`ALTER TABLE merkle_roots ADD COLUMN contract_address TEXT;`)

	fmt.Println("DB initialized and migrations applied")
	return db, nil
}

// CloseDB closes connection
func CloseDB(db *sql.DB) {
	if db != nil {
		db.Close()
	}
}
