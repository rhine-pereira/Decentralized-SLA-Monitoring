# Aggregator Service (Prototype)

This aggregator ingests watcher-node JSONL logs, groups them into 60-second buckets, builds Merkle trees, generates inclusion proofs, and exposes a REST API to fetch proofs.

Quickstart (local):
1. Ensure Go 1.20+ is installed.
2. Clone the repo and create aggregator/ with files from session workspace.
3. Run `go mod tidy` in aggregator/.
4. Build and run: `go run . -config config.yaml` (or `go build` then `./aggregator`).

API Endpoints:
- GET /proof?url=<url>&timestamp=<unix_ts>
- GET /root?bucket_id=<bucket>
- GET /health

Design notes:
- Leaf = SHA256(JSON(SignedRecord)) where SignedRecord contains data+signature+public_key
- Time window: 60 seconds (bucket_id = timestamp / 60)
- Storage: SQLite (raw_logs, merkle_roots) + JSON files in output/

Testing:
- Unit tests: `go test ./...`
- Integration: integration_test.go
- E2E: e2e_test_script.sh (bash)
