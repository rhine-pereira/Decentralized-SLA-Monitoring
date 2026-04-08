# watcher-node

Minimal watcher node MVP (Go) for decentralized SLA monitoring.

Usage:
1. Build: `go build` (run in watcher-node directory)
2. Run a node: `./watcher-node -config config_A.json`
3. Verify: `./watcher-node_verify -log logs/watcher_A.jsonl` (binary name depends on build)

Notes:
- Keys and logs are stored in `keys/` and `logs/` by default.
- Private keys are saved base64-encoded in `keys/node_<ID>.key` (Week 1 MVP).
- The verification CLI `verify_cli.go` reads a jsonl file and validates signatures.
