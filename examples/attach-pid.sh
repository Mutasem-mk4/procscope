#!/bin/bash
# Example: Attach to a running process
#
# This example demonstrates attaching to an existing PID.

set -e

echo "=== procscope: PID Attach Example ==="
echo ""

if [ -z "$1" ]; then
    echo "Usage: $0 <PID>"
    echo ""
    echo "Example: Start a long-running process, then:"
    echo "  sleep 300 &"
    echo "  $0 \$!"
    exit 1
fi

PID="$1"
echo "Attaching to PID $PID..."
echo "Press Ctrl+C to stop."
echo ""

sudo procscope -p "$PID"
