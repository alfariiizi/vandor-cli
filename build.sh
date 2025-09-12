#!/bin/bash

# Build script for Vandor CLI

set -e

echo "Building Vandor CLI..."

# Build the CLI binary
go build -o vandor main.go

echo "✅ Vandor CLI built successfully!"
echo "Run './vandor --help' to get started"