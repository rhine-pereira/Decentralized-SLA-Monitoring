package main

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "io/ioutil"

    "gopkg.in/yaml.v3"
)

// MonitoringResult represents the watcher data payload
type MonitoringResult struct {
    URL         string `json:"url"`
    Status      int    `json:"status"`
    LatencyMs   int64  `json:"latency_ms"`
    Timestamp   int64  `json:"timestamp"`
    ResponseHash string `json:"response_hash"`
    WatcherID   string `json:"watcher_id"`
    ErrorType   string `json:"error_type,omitempty"`
}

// SignedResult wraps monitoring data with signature and public key
type SignedResult struct {
    Data      MonitoringResult `json:"data"`
    Signature string           `json:"signature"`
    PublicKey string           `json:"public_key"`
}

// Config contains runtime configuration loaded from YAML
type Config struct {
    WatcherLogsDir string `yaml:"watcher_logs_dir"`
    TimeWindowSec  int    `yaml:"time_window_seconds"`
    DatabasePath   string `yaml:"database_path"`
    OutputDir      string `yaml:"output_dir"`
    PollInterval   int    `yaml:"poll_interval_seconds"`
    API            struct {
        Port int `yaml:"port"`
    } `yaml:"api"`
    Blockchain struct {
        Enabled         bool   `yaml:"enabled"`
        RPCURL          string `yaml:"rpc_url"`
        PrivateKey      string `yaml:"private_key"`
        ContractAddress string `yaml:"contract_address"`
    } `yaml:"blockchain"`
}

// LoadConfig reads YAML config from path
func LoadConfig(path string) (*Config, error) {
    b, err := ioutil.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var cfg Config
    if err := yaml.Unmarshal(b, &cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}

// ComputeLeafHash serializes the entire signed object and returns its SHA256 hex string
func ComputeLeafHash(rec SignedResult) (string, error) {
    b, err := json.Marshal(rec)
    if err != nil {
        return "", err
    }
    h := sha256.Sum256(b)
    return hex.EncodeToString(h[:]), nil
}
