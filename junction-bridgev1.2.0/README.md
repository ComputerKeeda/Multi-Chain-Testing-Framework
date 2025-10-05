# Junction Bridge Testing Tool

A comprehensive tool for setting up and managing Junction blockchain nodes for bridge testing operations.

## Features

- **Node Initialization**: Automatically initialize Junction blockchain nodes with custom configuration
- **Genesis Configuration**: Modify genesis files with appropriate voting and deposit periods
- **Key Management**: Generate and manage validator keys
- **Configuration Management**: Flexible configuration through YAML files and command-line flags
- **Governance Proposals**: Submit and manage governance proposals for EVM bridge parameter updates
- **Voting System**: Vote on governance proposals with validation
- **Proposal Monitoring**: Real-time monitoring of proposal status with animations
- **IPFS Integration**: Support for IPFS metadata uploads

## Quick Start

### 1. Build the Tool

```bash
# Make the build script executable
chmod +x build_executable.sh

# Build the executable
./build_executable.sh
```

### 2. Run Node Initialization

```bash
# Run with default configuration
./build/junction-bridge init-node

# Run with custom parameters
./build/junction-bridge init-node --moniker my-node --chain-id my-chain --key-name my-key
```

### 3. Submit Governance Proposals

```bash
# Submit a governance proposal (will prompt for IPFS CID)
./build/junction-bridge submit-proposal

# Vote on a proposal
./build/junction-bridge vote <proposal-id> <vote-option>

# Monitor proposal status
./build/junction-bridge monitor-proposals
```

## Configuration

The tool can be configured through:

1. **config.yaml** file (default configuration)
2. **Command-line flags** (override config file)
3. **Environment variables** (highest priority)

### Default Configuration

```yaml
moniker: "junction-testing"
chain_id: "junction"
denom: "uamf"
key_name: "test1"
amount: "100000000000uamf"
validator_stake: "10000000000uamf"
junctiond_path: "./build/junctiond"
home_dir: "$HOME/.junction"
minimum_gas_prices: "0.00025uamf"
rest_endpoint: "http://localhost:1317"
```

## What the Tool Does

### Node Initialization (`init-node`)

1. **Cleans Environment**: Removes existing junctiond directory
2. **Initializes Node**: Creates new blockchain node with specified parameters
3. **Generates Keys**: Creates validator keys for the node
4. **Sets Up Genesis**: Adds genesis account and creates gentx
5. **Configures Governance**: Updates voting and deposit periods
6. **Starts Node**: Launches the blockchain node with proper gas settings

### Governance Operations (`submit-proposal`, `vote`, `monitor-proposals`)

1. **Metadata Creation**: Creates metadata.json from draft template
2. **IPFS Upload Guidance**: Provides instructions for uploading to IPFS
3. **Proposal Creation**: Generates proposal.json with EVM bridge parameter updates using IPFS CID
4. **Proposal Submission**: Submits governance proposal to the blockchain
5. **Voting**: Allows voting on proposals with validation
6. **Monitoring**: Real-time proposal status monitoring with animations
7. **Completion Detection**: Shows completion animation when voting period ends

## Command Line Options

### Node Initialization

```bash
./build/junction-bridge init-node [flags]

Flags:
  --amount string              Initial amount (default "100000000000uamf")
  --chain-id string            Chain ID (default "junction")
  --denom string               Denomination (default "uamf")
  --home-dir string            Home directory (default "$HOME/.junction")
  --junctiond-path string      Path to junctiond binary (default "./build/junctiond")
  --key-name string            Key name (default "test1")
  --minimum-gas-prices string  Minimum gas prices (default "0.00025uamf")
  --moniker string             Moniker for the node (default "junction-testing")
  --validator-stake string     Validator stake amount (default "10000000000uamf")
```

### Governance Operations

```bash
# Submit a governance proposal
./build/junction-bridge submit-proposal

# Vote on a proposal
./build/junction-bridge vote <proposal-id> <vote-option>
# Vote options: yes, no, abstain, no_with_veto

# Monitor proposal status
./build/junction-bridge monitor-proposals
```

## Requirements

- Go 1.21 or higher
- **Junction blockchain binary (`junctiond`)** - Must be available for blockchain operations
- Sufficient disk space for blockchain data

### Junction Binary Setup

The tool requires the `junctiond` blockchain binary to perform operations like:

- `junctiond tx gov submit-proposal`
- `junctiond tx gov vote`
- `junctiond keys add`
- `junctiond init`

**Option 1: Automatic download (recommended)**

```bash
# The build script will automatically offer to download junctiond
./build_executable.sh
# Choose option 1 when prompted
```

**Option 2: Manual download**

```bash
# Download from GitHub release
curl -L -o ./build/junctiond https://github.com/ComputerKeeda/junction/releases/download/bridge-v1.2.0/junctiond
chmod +x ./build/junctiond
```

**Option 3: Use your own binary**

```bash
# Copy your junctiond binary to:
cp /path/to/your/junctiond ./build/junctiond
chmod +x ./build/junctiond
```

**Option 4: Specify custom path**

```yaml
# In config.yaml
junctiond_path: "/path/to/your/junctiond"
```

## File Structure

```
junction-bridgev1.2.0/
├── main.go                 # Main application code
├── go.mod                  # Go module definition
├── config.yaml            # Default configuration
├── build_executable.sh    # Build script
├── README.md              # This file
├── draft_metadata.json    # Draft metadata template
├── draft_proposal.json    # Draft proposal template
└── build/                 # Build output directory
    ├── junction-bridge    # Our compiled executable
    └── junctiond          # Junction blockchain binary (required)
```

**Generated Files (during runtime):**

- `metadata.json` - Created from draft template
- `proposal.json` - Created with IPFS CID
- `$HOME/.junction/` - Blockchain data directory

## Troubleshooting

### Common Issues

1. **Permission Denied**: Make sure the build script is executable

   ```bash
   chmod +x build_executable.sh
   ```

2. **Junctiond Not Found**: Ensure the `junctiond` binary is in `./build/junctiond`

   ```bash
   # Check if junctiond exists
   ls -la ./build/junctiond
   ```

3. **Port Already in Use**: The default port might be in use, check with:
   ```bash
   netstat -tulpn | grep :26657
   ```

## Development

To modify the tool:

1. Edit `main.go` for functionality changes
2. Update `config.yaml` for default configuration changes
3. Modify `build_executable.sh` for build process changes
4. Rebuild with `./build_executable.sh`

## License

This project is part of the Junction Bridge testing infrastructure.
