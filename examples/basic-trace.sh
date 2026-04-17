#!/bin/bash
# Example: Basic process tracing with procscope
#
# This example traces a simple command and shows the live timeline.

set -e

echo "=== procscope: Basic Trace Example ==="
echo ""
echo "Tracing 'ls -la /tmp' with procscope..."
echo ""

sudo procscope -- ls -la /tmp

echo ""
echo "=== Done ==="
