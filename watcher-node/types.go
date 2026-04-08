package main

import "time"

// MonitoringResult is the canonical monitoring data structure. Field order is important for deterministic JSON.
type MonitoringResult struct {
	URL         string `json:"url"`
	Status      int    `json:"status"`
	LatencyMs   int64  `json:"latency_ms"`
	Timestamp   int64  `json:"timestamp"`
	ResponseHash string `json:"response_hash"`
	WatcherID   string `json:"watcher_id"`
	ErrorType   string `json:"error_type,omitempty"`
}

// SignedResult wraps the monitoring data with signature and public key.
type SignedResult struct {
	Data      MonitoringResult `json:"data"`
	Signature string           `json:"signature"`
	PublicKey string           `json:"public_key"`
}

// NodeConfig represents the local node configuration.
type NodeConfig struct {
	NodeID         string   `json:"node_id"`
	URLs           []string `json:"urls"`
	IntervalSec    int      `json:"interval_seconds"`
	LogDir         string   `json:"log_dir"`
	KeyDir         string   `json:"key_dir"`
}

// simple helper for timestamping
func nowUnix() int64 { return time.Now().Unix() }
