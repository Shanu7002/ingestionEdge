#!/bin/bash

echo "Initiating HTTP Volumetric Flood..."
echo "Target: http://127.0.0.1:8080/ingest"
echo "Concurrency: 50 threads | Total Requests: 1,000"
echo "---------------------------------------------------"

echo '{"alert": "OOM_Warning", "severity": "P0"}' > /tmp/payload.json

# - seq 1 1000     : Generates numbers 1 to 1000
# - xargs -P 50    : Spawns 50 parallel worker processes
# - curl -w        : Extracts only the HTTP status code (e.g., 202 or 429)
# - sort | uniq -c : Aggregates and counts the exact results

seq 1 1000 | xargs -n1 -P 50 -I {} \
  curl -s -w "%{http_code}\n" -o /dev/null \
  -X POST -H "Content-Type: application/json" \
  -d @/tmp/payload.json \
  http://127.0.0.1:8080/ingest \
  | sort | uniq -c

echo "---------------------------------------------------"
echo "Test complete. Cleaning up."
rm /tmp/payload.json