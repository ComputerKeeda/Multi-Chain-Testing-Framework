#!/bin/bash

# Setup script for junctiond binary
echo "🔧 Junction Binary Setup Helper"
echo "==============================="

# Check if junctiond already exists
if [ -f "./build/junctiond" ]; then
    echo "✅ junctiond binary already exists at ./build/junctiond"
    exit 0
fi

echo "📋 This script helps you set up the junctiond binary required for blockchain operations."
echo ""
echo "The tool needs junctiond to perform operations like:"
echo "  - junctiond tx gov submit-proposal"
echo "  - junctiond tx gov vote"
echo "  - junctiond keys add"
echo "  - junctiond init"
echo ""

echo "🔧 Options to get junctiond binary:"
echo "1. Download from GitHub release (recommended)"
echo "2. Use your own binary"
echo ""
read -p "Choose option (1/2): " choice

case $choice in
    1)
        echo "📥 Downloading junctiond from GitHub release..."
        echo "🔗 URL: https://github.com/ComputerKeeda/junction/releases/download/bridge-v1.2.0/junctiond"
        
        # Download using curl or wget
        if command -v curl >/dev/null 2>&1; then
            curl -L -o ./build/junctiond https://github.com/ComputerKeeda/junction/releases/download/bridge-v1.2.0/junctiond
        elif command -v wget >/dev/null 2>&1; then
            wget -O ./build/junctiond https://github.com/ComputerKeeda/junction/releases/download/bridge-v1.2.0/junctiond
        else
            echo "❌ Error: Neither curl nor wget found. Please install one of them or download manually."
            exit 1
        fi
        
        if [ $? -eq 0 ]; then
            chmod +x ./build/junctiond
            echo "✅ junctiond downloaded and made executable!"
        else
            echo "❌ Error: Failed to download junctiond binary"
            exit 1
        fi
        ;;
    2)
        echo "🔍 Please provide the path to your junctiond binary:"
        echo "   (This should be the actual junctiond executable, not a directory)"
        read -p "Path to junctiond: " junctiond_path
        
        if [ ! -f "$junctiond_path" ]; then
            echo "❌ Error: File not found at $junctiond_path"
            exit 1
        fi
        
        cp "$junctiond_path" ./build/junctiond
        chmod +x ./build/junctiond
        echo "✅ junctiond copied and made executable!"
        ;;
    *)
        echo "❌ Invalid choice. Exiting."
        exit 1
        ;;
esac

echo "✅ junctiond binary setup complete!"
echo "📍 Location: ./build/junctiond"
echo ""
echo "🚀 You can now run:"
echo "  ./build/junction-bridge init-node"
echo "  ./build/junction-bridge submit-proposal"
