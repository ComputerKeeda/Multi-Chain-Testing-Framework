#!/bin/bash

# Build script for Junction Bridge Testing Tool
echo "🔨 Building Junction Bridge Testing Tool..."

# Create build directory if it doesn't exist
mkdir -p build

# Build the main executable
echo "📦 Building main executable..."
go build -o build/junction-bridge main.go

# Make it executable
chmod +x build/junction-bridge

echo "✅ Build completed!"
echo "📁 Executable created at: ./build/junction-bridge"
echo ""

# Check if junctiond binary exists
if [ ! -f "./build/junctiond" ]; then
    echo "⚠️  junctiond binary not found at ./build/junctiond"
    echo ""
    echo "🔧 Options to get junctiond binary:"
    echo "1. Download from GitHub release (recommended)"
    echo "2. Use your own binary"
    echo "3. Skip (specify custom path in config.yaml)"
    echo ""
    read -p "Choose option (1/2/3): " choice
    
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
            echo "📁 Please provide the path to your junctiond binary:"
            read -p "Path to junctiond: " junctiond_path
            
            if [ ! -f "$junctiond_path" ]; then
                echo "❌ Error: File not found at $junctiond_path"
                exit 1
            fi
            
            cp "$junctiond_path" ./build/junctiond
            chmod +x ./build/junctiond
            echo "✅ junctiond copied and made executable!"
            ;;
        3)
            echo "⏭️  Skipping junctiond setup"
            echo "💡 You can specify a custom path in config.yaml:"
            echo "   junctiond_path: \"/path/to/your/junctiond\""
            ;;
        *)
            echo "❌ Invalid choice. Exiting."
            exit 1
            ;;
    esac
else
    echo "✅ junctiond binary found at ./build/junctiond"
fi

echo ""
echo "🚀 Usage:"
echo "  ./build/junction-bridge init-node --help"
echo "  ./build/junction-bridge init-node"
echo ""
echo "💡 You can also use custom parameters:"
echo "  ./build/junction-bridge init-node --moniker my-node --chain-id my-chain"
