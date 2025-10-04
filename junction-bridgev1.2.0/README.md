# Junction Chain Testing Scripts

A comprehensive Go-based executable for testing Junction blockchain with interactive CLI, parameter change proposals, and automated voting mechanisms.

## Features

🚀 **Interactive Chain Setup**

- Automated chain initialization with configurable parameters
- Genesis account creation and validator staking
- Genesis file modification with custom voting periods

🔧 **Parameter Change Proposals**

- Interactive parameter collection for bridge workers and contract addresses
- Environment variable support for automated configuration
- JSON proposal file generation

🗳️ **Voting & Governance**

- Automated proposal submission
- Interactive voting with multiple options
- Voting period countdown with real-time animations

⏰ **CLI Animations**

- Loading spinners for all operations
- Countdown timers for voting periods
- Progress indicators for long-running tasks

## Quick Start

### 1. Build the Executable

```bash
# Make the build script executable
chmod +x build_executable.sh

# Build the Go executable
./build_executable.sh
```

### 2. Run the Chain Tester

```bash
# Run the interactive chain tester
./chain-tester
```

## Configuration

### Environment Variables

Create a `.env` file (copy from `env.example`) to configure:

```bash
# Chain Configuration
MONIKER=junction-testing
CHAIN_ID=junction
DENOM=uamf
KEY_NAME=test1
AMOUNT=100000000000uamf
VALIDATOR_STAKE=10000000000uamf
GAS_PRICES=0.0025uamf
MINIMUM_GAS_PRICES=0.00025uamf

# Bridge Configuration
BRIDGE_WORKERS=air1abc...,air1def...,air1ghi...
BRIDGE_CONTRACT_ADDRESS=0x1234567890123456789012345678901234567890

# Proposal Configuration
PROPOSAL_TITLE=Update EVM Bridge Authorized Unlockers
PROPOSAL_DESCRIPTION=Add new addresses to the authorized unlockers list
PROPOSAL_DEPOSIT=1000000uamf
PROPOSER_KEY=test1

# Voting Configuration
VOTING_PERIOD=600
VOTE_OPTION=yes
```

### Interactive Mode

If environment variables are not set, the script will prompt for:

- Bridge worker addresses (comma-separated)
- Bridge contract address
- Proposal ID for voting
- Vote options (yes/no/no_with_veto/abstain)
- Voting period duration

## Usage Examples

### Basic Chain Setup

```bash
# Run with default configuration
./chain-tester
```

### With Environment Variables

```bash
# Set environment variables
export BRIDGE_WORKERS="air1abc123,air1def456,air1ghi789"
export BRIDGE_CONTRACT_ADDRESS="0x1234567890123456789012345678901234567890"
export VOTE_OPTION="yes"

# Run the tester
./chain-tester
```

### Using .env File

```bash
# Copy example environment file
cp env.example .env

# Edit .env with your values
nano .env

# Run the tester
./chain-tester
```

## Generated Files

The script creates several files during execution:

- `proposal.json` - Parameter change proposal file
- `~/.junction/` - Chain data directory
- `~/.junction/config/genesis.json` - Modified genesis file

## Commands Executed

The script automatically executes these commands in sequence:

1. **Cleanup**: `rm -rf ~/.junction`
2. **Initialize**: `junctiond init junction-testing --default-denom uamf --chain-id junction`
3. **Generate Keys**: `junctiond keys add test1 --keyring-backend os`
4. **Add Account**: `junctiond genesis add-genesis-account test1 100000000000uamf --keyring-backend os`
5. **Stake Validator**: `junctiond genesis gentx test1 10000000000uamf --keyring-backend os --gas-prices 0.0025uamf --chain-id junction`
6. **Collect Gentx**: `junctiond genesis collect-gentxs`
7. **Modify Genesis**: `jq` commands to update voting periods
8. **Submit Proposal**: `junctiond tx gov submit-proposal param-change proposal.json --from test1 --deposit 1000000uamf`
9. **Vote**: `junctiond tx gov vote <proposal_id> yes --from test1 --keyring-backend os --chain-id junction`
10. **Start Node**: `junctiond start --minimum-gas-prices 0.00025uamf`

## Requirements

- Go 1.21 or later
- `junctiond` binary in `./build/junctiond`
- `jq` for JSON processing
- Proper permissions to execute binaries

## Troubleshooting

### Common Issues

1. **Permission Denied**: Make sure the executable has proper permissions

   ```bash
   chmod +x chain-tester
   ```

2. **Missing junctiond**: Ensure `junctiond` binary exists in `./build/junctiond`

   ```bash
   ls -la build/junctiond
   ```

3. **Missing jq**: Install jq for JSON processing

   ```bash
   # Ubuntu/Debian
   sudo apt-get install jq

   # macOS
   brew install jq
   ```

4. **Environment Variables Not Loading**: Check `.env` file format
   ```bash
   # Ensure no spaces around =
   KEY=value
   # Not: KEY = value
   ```

## Development

### Building from Source

```bash
# Install dependencies
go mod tidy

# Build the executable
go build -o chain-tester main.go

# Run tests (if any)
go test
```

### Customizing the Script

The script is modular and can be easily extended:

- Add new environment variables in `loadConfig()`
- Modify proposal structure in `Proposal` and `BridgeParams` types
- Add new CLI animations in `showLoadingAnimation()`
- Extend voting options in `voteOnProposal()`

## License

This project is part of the Junction blockchain testing suite.
