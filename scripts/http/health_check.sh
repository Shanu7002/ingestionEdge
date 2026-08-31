#!/bin/bash

echo "Health check started"
echo "Target: http://127.0.0.1:8080/health"
echo "---------------------------------------------------"

# - curl -w        : Extracts only the HTTP status code (e.g., 202 or 429)

curl -s -w "%{http_code}\n" -o /dev/null \
-X GET -H "Content-Type: application/json" \
http://127.0.0.1:8080/health \
| jq

echo "---------------------------------------------------"
echo "Test complete. Cleaning up."