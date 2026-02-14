#!/bin/bash
# Build script for DuckDB-enabled Lambda handlers
# This script runs inside the AL2023 Docker container
# to ensure glibc compatibility with Lambda runtime.

set -e

echo "========================================================"
echo "DuckDB Handler Build (AL2023 x86_64)"
echo "========================================================"

HANDLER=${1:-data-explorer}
OUTPUT_DIR=${2:-services/api/data-explorer/.bin}

echo "Handler:    $HANDLER"
echo "Output:     $OUTPUT_DIR"
echo "Go Version: $(go version)"
echo ""

cd /workspace

mkdir -p "$OUTPUT_DIR"

echo "Building $HANDLER handler with CGO..."
cd "cmd/handlers/$HANDLER"

CGO_ENABLED=1 \
GOOS=linux \
GOARCH=amd64 \
go build -ldflags="-s -w" -o "/workspace/$OUTPUT_DIR/bootstrap" .

echo ""
echo "Build complete: $OUTPUT_DIR/bootstrap"

ls -lh "/workspace/$OUTPUT_DIR/bootstrap"
file "/workspace/$OUTPUT_DIR/bootstrap"
echo "========================================================"
