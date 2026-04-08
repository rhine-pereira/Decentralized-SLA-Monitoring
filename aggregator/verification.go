package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
)

// VerifySignedResult verifies the ed25519 signature over the marshaled Data field.
// Returns nil if signature is valid.
func VerifySignedResult(s SignedResult) error {
	b, err := json.Marshal(s.Data)
	if err != nil {
		return err
	}
	sig, err := base64.StdEncoding.DecodeString(s.Signature)
	if err != nil {
		return err
	}
	pubBytes, err := hex.DecodeString(s.PublicKey)
	if err != nil {
		return err
	}
	if len(pubBytes) != ed25519.PublicKeySize {
		return errors.New("invalid public key size")
	}
	if !ed25519.Verify(ed25519.PublicKey(pubBytes), b, sig) {
		return errors.New("signature verification failed")
	}
	return nil
}
