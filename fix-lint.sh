#!/bin/bash
set -e

echo "🔧 Fixing linting issues..."

echo "📝 Running gofmt..."
gofmt -w .

echo "📝 Running goimports..."
goimports -w .

echo "🧹 Running golangci-lint with auto-fix for critical errors..."
golangci-lint run --fix --disable=revive,unused-parameter --timeout=5m

echo "✅ Linting fixes complete!"

echo "🧪 Running tests to ensure nothing is broken..."
go test ./...

echo "🏗️ Testing build..."
go build -o /tmp/vandor-test main.go && rm -f /tmp/vandor-test

echo "✅ All done! Your code should now pass CI."