#!/bin/bash

echo "🔨 Building Junction Chain Testing Executable"
echo "============================================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go first."
    echo "Visit: https://golang.org/doc/install"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | cut -d' ' -f3 | cut -d'o' -f2)
echo "📋 Go version: $GO_VERSION"

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -f chain-tester
rm -f chain-tester.exe

# Build the executable
echo "🔨 Building executable..."
go build -o chain-tester main.go

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
    echo "📦 Executable created: chain-tester"
    echo ""
    echo "🚀 To run the executable:"
    echo "   ./chain-tester"
    echo ""
    echo "📝 Make sure you have:"
    echo "   - junctiond binary in ./build/junctiond"
    echo "   - jq installed for JSON processing"
    echo "   - Proper permissions to execute the binary"
    
    # Make the executable executable
    chmod +x chain-tester
    
    echo ""
    echo "🎯 Ready to test your chain!"
else
    echo "❌ Build failed!"
    exit 1
fi
