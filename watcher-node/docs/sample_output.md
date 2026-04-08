# Sample Signed Output

Example line from watcher-node/sample_watcher_A.jsonl (JSON Lines):

{"data":{"url":"https://example.com/","status":200,"latency_ms":123,"timestamp":1710000000,"response_hash":"abcdef012345...","watcher_id":"A"},"signature":"BASE64_SIGNATURE_EXAMPLE","public_key":"HEX_PUBKEY_EXAMPLE"}

Note: The real system will produce valid Ed25519 signatures and hex-encoded public keys. Use the verify CLI to validate entries.
