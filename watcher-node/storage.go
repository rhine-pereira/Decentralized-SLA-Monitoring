package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// WriteSigned appends a SignedResult as a single JSON line to the log file.
func WriteSigned(logDir, nodeID string, s SignedResult) error {
	_ = os.MkdirAll(logDir, 0o700)
	path := filepath.Join(logDir, "watcher_"+nodeID+".jsonl")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil { return err }
	defer f.Close()
	b, err := json.Marshal(s)
	if err != nil { return err }
	if _, err := f.Write(append(b, '\n')); err != nil { return err }
	return f.Sync()
}
