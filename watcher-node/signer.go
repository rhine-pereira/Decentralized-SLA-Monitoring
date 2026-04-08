package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

// GenerateOrLoadKey loads a base64 private key if present, otherwise generates and saves one.
func GenerateOrLoadKey(keyDir, nodeID string) (ed25519.PublicKey, ed25519.PrivateKey, error) {
	_ = os.MkdirAll(keyDir, 0o700)
	keyPath := filepath.Join(keyDir, "node_"+nodeID+".key")
	if _, err := os.Stat(keyPath); err == nil {
		b, err := ioutil.ReadFile(keyPath)
		if err != nil { return nil, nil, err }
		privBytes, err := base64.StdEncoding.DecodeString(string(b))
		if err != nil { return nil, nil, err }
		if len(privBytes) != ed25519.PrivateKeySize { return nil, nil, errors.New("invalid private key size") }
		priv := ed25519.PrivateKey(privBytes)
		pub := ed25519.PublicKey(priv[32:])
		return pub, priv, nil
	}
	// generate
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil { return nil, nil, err }
	// save base64 of private key
	enc := base64.StdEncoding.EncodeToString([]byte(priv))
	if err := ioutil.WriteFile(keyPath, []byte(enc), 0o600); err != nil { return nil, nil, err }
	return pub, priv, nil
}

// SignResult marshals the MonitoringResult deterministically and signs it.
func SignResult(priv ed25519.PrivateKey, data MonitoringResult) (SignedResult, error) {
	// marshal data (struct field order ensures determinism)
	b, err := json.Marshal(data)
	if err != nil { return SignedResult{}, err }
	sig := ed25519.Sign(priv, b)
	pub := priv[32:]
	return SignedResult{
		Data:      data,
		Signature: base64.StdEncoding.EncodeToString(sig),
		PublicKey: hex.EncodeToString(pub),
	}, nil
}

// VerifySigned checks the signature in a SignedResult. Returns nil on success.
func VerifySigned(s SignedResult) error {
	// reconstruct bytes of data
	b, err := json.Marshal(s.Data)
	if err != nil { return err }
	sig, err := base64.StdEncoding.DecodeString(s.Signature)
	if err != nil { return err }
	pubBytes, err := hex.DecodeString(s.PublicKey)
	if err != nil { return err }
	if len(pubBytes) != ed25519.PublicKeySize { return errors.New("invalid public key size") }
	if !ed25519.Verify(ed25519.PublicKey(pubBytes), b, sig) {
		return errors.New("signature verification failed")
	}
	return nil
}
