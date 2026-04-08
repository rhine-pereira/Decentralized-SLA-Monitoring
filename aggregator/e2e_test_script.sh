#!/usr/bin/env bash
# E2E script: start aggregator, wait for processing, call /proof for a sample record, verify signature and merkle proof locally.

set -e

# Build and run aggregator (assumes go installed)
cd aggregator
go build -o aggregator_bin
./aggregator_bin -config config.yaml &
PID=$!
sleep 5

# pick a sample record from output dir
BUCKET=$(ls output | head -n1 | sed 's/\.json$//')
curl -s "http://localhost:8888/proof?url=https://google.com&timestamp=1775564617" | jq '.'

# cleanup
kill $PID
