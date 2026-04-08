package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
)

// VerifyLine verifies a single JSON line (SignedResult) and returns nil on success.
func VerifyLine(line string) error {
	var s SignedResult
	if err := json.Unmarshal([]byte(line), &s); err != nil { return err }
	if s.PublicKey == "" || s.Signature == "" { return errors.New("missing signature or pubkey") }
	return VerifySigned(s)
}

// VerifyFile reads a .jsonl file and verifies all entries, returning counts.
func VerifyFile(path string) (total, valid, invalid int, err error) {
	f, err := os.Open(path)
	if err != nil { return }
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		total++
		if err := VerifyLine(s.Text()); err != nil { invalid++ } else { valid++ }
	}
	if s.Err() != nil { err = s.Err() }
	return
}
