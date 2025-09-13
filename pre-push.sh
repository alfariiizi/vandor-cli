#!/bin/bash
set -e

echo "🚀 Pre-push checks for vandor-cli"
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print status
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if we're in the right directory
if [ ! -f "main.go" ] || [ ! -f "go.mod" ]; then
    print_error "Not in the vandor-cli root directory!"
    exit 1
fi

echo "📁 Working directory: $(pwd)"
echo

# 1. Check Go module is tidy
echo "🧹 Checking if go.mod is tidy..."
go mod tidy
if git diff --exit-code go.mod go.sum; then
    print_status "go.mod is tidy"
else
    print_warning "go.mod was not tidy - fixed automatically"
fi
echo

# 2. Format code
echo "🎨 Formatting code..."
gofmt -w .
goimports -w .
print_status "Code formatted"
echo

# 3. Run tests
echo "🧪 Running tests..."
if go test -race -v ./...; then
    print_status "All tests pass"
else
    print_error "Tests failed!"
    exit 1
fi
echo

# 4. Build project
echo "🏗️  Building project..."
if go build -o /tmp/vandor-test main.go; then
    rm -f /tmp/vandor-test
    print_status "Build successful"
else
    print_error "Build failed!"
    exit 1
fi
echo

# 5. Run critical linting (CI-friendly)
echo "🔍 Running critical linting checks..."
if golangci-lint run --disable=revive,unused-parameter --timeout=5m; then
    print_status "Critical linting passed"
else
    print_error "Critical linting failed! Run './fix-lint.sh' to fix automatically."
    exit 1
fi
echo

# 6. Security check (if gosec is available)
if command -v gosec >/dev/null 2>&1; then
    echo "🔒 Running security check..."
    if gosec -quiet ./...; then
        print_status "Security check passed"
    else
        print_warning "Security issues found - please review"
    fi
    echo
fi

# 7. Check for common issues
echo "🔍 Checking for common issues..."

# Check for TODO/FIXME comments
if grep -r "TODO\|FIXME" --include="*.go" .; then
    print_warning "Found TODO/FIXME comments - consider addressing before push"
else
    print_status "No TODO/FIXME comments found"
fi
echo

# Check for debug prints
if grep -r "fmt.Print\|log.Print\|spew.Dump" --include="*.go" . | grep -v "_test.go" | grep -v "// allowed"; then
    print_warning "Found debug prints - consider removing before push"
else
    print_status "No debug prints found"
fi
echo

# 8. Check git status
echo "📋 Git status check..."
if git diff --quiet && git diff --staged --quiet; then
    print_status "No uncommitted changes"
else
    print_warning "You have uncommitted changes:"
    git status --short
    echo
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_error "Aborted by user"
        exit 1
    fi
fi

# 9. Final summary
echo "🎉 Pre-push checks complete!"
echo "=================================="
print_status "✅ Go module is tidy"
print_status "✅ Code is formatted"
print_status "✅ All tests pass"
print_status "✅ Build successful"
print_status "✅ Critical linting passed"
echo
echo "🚀 Ready to push! Your code should pass CI."
echo
echo "Useful commands:"
echo "  git push origin main    # Push to main branch"
echo "  ./fix-lint.sh          # Fix all linting issues"
echo "  golangci-lint run      # Full lint check"