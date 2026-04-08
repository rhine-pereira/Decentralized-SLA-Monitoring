package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	configPath := flag.String("config", "config_A.json", "Path to node config JSON")
	flag.Parse()
	b, err := ioutil.ReadFile(*configPath)
	if err != nil { log.Fatalf("failed to read config: %v", err) }
	var cfg NodeConfig
	if err := json.Unmarshal(b, &cfg); err != nil { log.Fatalf("invalid config: %v", err) }
	if cfg.IntervalSec <= 0 { cfg.IntervalSec = 5 }
	if cfg.LogDir == "" { cfg.LogDir = "logs" }
	if cfg.KeyDir == "" { cfg.KeyDir = "keys" }
	// ensure directories exist
	_ = os.MkdirAll(cfg.LogDir, 0o700)
	_ = os.MkdirAll(cfg.KeyDir, 0o700)
	pub, priv, err := GenerateOrLoadKey(cfg.KeyDir, cfg.NodeID)
	if err != nil { log.Fatalf("key error: %v", err) }
	_ = pub // unused directly; signer includes public in SignedResult
	interval := time.Duration(cfg.IntervalSec) * time.Second
	log.Printf("starting watcher %s: %d urls, interval=%ds", cfg.NodeID, len(cfg.URLs), cfg.IntervalSec)

	// Graceful shutdown context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		// Run one monitoring pass for all URLs
		for _, url := range cfg.URLs {
			select {
			case <-ctx.Done():
				log.Printf("shutdown requested, exiting monitor loop")
				return
			default:
				res := MonitorURL(url, cfg.NodeID, 10*time.Second)
				signed, err := SignResult(priv, res)
				if err != nil { fmt.Printf("sign error: %v\n", err); continue }
				if err := WriteSigned(cfg.LogDir, cfg.NodeID, signed); err != nil { fmt.Printf("write error: %v\n", err) }
			}
		}

		// Wait for next tick or shutdown
		select {
		case <-ctx.Done():
			log.Printf("shutdown requested, exiting")
			return
		case <-ticker.C:
			// continue
		}
	}
}
