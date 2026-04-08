package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/local/decentralized-sla/aggregator/chain"
)

func main_impl() {
	cfgPath := flag.String("config", "config.yaml", "path to config yaml")
	flag.Parse()
	cfg, err := LoadConfig(*cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	// init DB
	db, err := InitDB(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("db init: %v", err)
	}
	defer CloseDB(db)

	var chainClient *chain.Client
	if cfg.Blockchain.Enabled {
		cc, err := chain.NewClient(cfg.Blockchain.RPCURL, cfg.Blockchain.PrivateKey, cfg.Blockchain.ContractAddress)
		if err != nil {
			log.Printf("failed to init blockchain client: %v", err)
		} else {
			chainClient = cc
			fmt.Println("Blockchain anchoring enabled")
		}
	}

	// initial parse of watcher logs
	n, err := parseWatcherLogs(db, cfg.WatcherLogsDir)
	if err != nil {
		log.Printf("loader error: %v", err)
	} else {
		fmt.Printf("Inserted %d records from logs\n", n)
	}

	// build merkle for all buckets present
	rows, err := db.Query("SELECT DISTINCT bucket_id FROM raw_logs ORDER BY bucket_id ASC")
	if err != nil {
		log.Fatalf("select buckets: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var b int64
		rows.Scan(&b)
		if err := BuildMerkleForBucket(db, b, cfg.OutputDir, chainClient, cfg.Blockchain.ContractAddress); err != nil {
			log.Printf("build bucket %d err: %v", b, err)
		}
	}

	// start API server
	go func() {
		if err := startAPI(db, chainClient, cfg.OutputDir, cfg.API.Port, cfg.Blockchain.ContractAddress); err != nil {
			log.Fatalf("api error: %v", err)
		}
	}()

	// simple loop to poll for new logs every poll interval
	for {
		time.Sleep(time.Duration(cfg.PollInterval) * time.Second)
		n, _ := parseWatcherLogs(db, cfg.WatcherLogsDir)
		if n > 0 {
			fmt.Printf("Inserted %d new records\n", n)
			// find new buckets and process them
			rows2, _ := db.Query("SELECT DISTINCT bucket_id FROM raw_logs ORDER BY bucket_id ASC")
			for rows2.Next() {
				var b int64
				rows2.Scan(&b)
				BuildMerkleForBucket(db, b, cfg.OutputDir, chainClient, cfg.Blockchain.ContractAddress)
			}
			rows2.Close()
		}
	}
}

// entrypoint used to avoid clashing with previously created main.go in session files
func main() {
	main_impl()
}
