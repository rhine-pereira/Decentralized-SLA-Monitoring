package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net"
	"net/http"
	"time"
)

// MonitorURL performs an HTTP GET, measures latency, and returns MonitoringResult.
func MonitorURL(url, watcherID string, timeout time.Duration) MonitoringResult {
	start := time.Now()
	client := &http.Client{Timeout: timeout}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "DecentralizedSLAWatcher/1.0")
	resp, err := client.Do(req)
	lat := time.Since(start)
	res := MonitoringResult{
		URL:       url,
		LatencyMs: lat.Milliseconds(),
		Timestamp: nowUnix(),
		WatcherID: watcherID,
	}
	if err != nil {
		// classify some errors
		switch err.(type) {
		case net.Error:
			res.ErrorType = "timeout_or_network"
		default:
			res.ErrorType = "network_error"
		}
		res.Status = 0
		res.ResponseHash = ""
		return res
	}
	defer resp.Body.Close()
	res.Status = resp.StatusCode
	// read body up to a limit
	limit := int64(1_000_000) // 1MB
	reader := io.LimitReader(resp.Body, limit)
	body, _ := io.ReadAll(reader)
	h := sha256.Sum256(body)
	res.ResponseHash = hex.EncodeToString(h[:])
	return res
}
