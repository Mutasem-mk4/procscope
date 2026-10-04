#!/bin/bash
# Example: Generate an evidence bundle
#
# This example traces a command and saves a full evidence bundle
# with JSONL events, process tree, and Markdown summary.

set -e

CASE_DIR="case-$(date +%Y%m%d-%H%M%S)"

echo "=== procscope: Evidence Bundle Example ==="
echo ""
echo "Tracing 'bash -c \"echo hello; ls /tmp; curl -s http://example.com > /dev/null\"'"
echo "Output directory: $CASE_DIR"
echo ""

sudo procscope \
    --out "$CASE_DIR" \
    --jsonl "$CASE_DIR/events.jsonl" \
    --summary "$CASE_DIR/report.md" \
    -- bash -c 'echo hello; ls /tmp; curl -s http://example.com > /dev/null 2>&1 || true'

echo ""
echo "=== Evidence Bundle Contents ==="
ls -la "$CASE_DIR/"

echo ""
echo "=== Summary Report ==="
cat "$CASE_DIR/report.md"
