#!/bin/bash

echo "Initiating UDP Volumetric Flood..."
echo "Target: 127.0.0.1:8125"
echo "Payload: 20,000 StatsD metrics"

# subshell
(
  for i in {1..20000}; do
    echo -n "cpu.load.$i:99|g" > /dev/udp/127.0.0.1/8125
  done
)

echo "Flood complete. Check your Go Edge terminal for load shedding logs."