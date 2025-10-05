# Junction Chain Testing Scripts

A comprehensive Go-based executable for testing Junction blockchain with **two-phase execution**, interactive CLI, parameter change proposals, and automated voting mechanisms.

## 🎯 **Two-Phase Testing System**

This chain tester uses a **smart two-phase approach** to handle the requirement that proposals can only be submitted when the chain is running:

### **Phase 1: Chain Setup & Proposal Creation**

- ✅ Automated chain initialization with configurable parameters
- ✅ Genesis account creation and validator staking
- ✅ Genesis file modification with custom voting periods
- ✅ Interactive parameter collection for bridge workers and contract addresses
- ✅ JSON proposal file generation
- ✅ **State persistence** - saves all user inputs for Phase 2

### **Phase 2: Chain Running + Proposal Submission**

- ✅ **Automatic chain startup** in background
- ✅ **New terminal opening** for proposal submission
- ✅ **State restoration** - loads all previous user inputs
- ✅ Automated proposal submission to running chain
- ✅ Interactive voting with multiple options
- ✅ Voting period countdown with real-time animations

## 🚀 **Key Features**

### **Smart State Management**

- **Persistent state** across phases using `testing_state.json`
- **User input preservation** - no need to re-enter bridge workers/contracts
- **Automatic cleanup** after completion

### **Seamless User Experience**

- **10-second countdown** before starting chain
- **Automatic terminal opening** for proposal submission
- **Clear phase indicators** throughout the process
- **No manual intervention** required between phases

### **CLI Animations & UX**

- Loading spinners for all operations
- Countdown timers for voting periods
- Progress indicators for long-running tasks
- Real-time status updates

## 🚀 **Quick Start**

### 1. Build the Executable

```bash
# Make the build script executable
chmod +x build_executable.sh

# Build the Go executable
./build_executable.sh
```

### 2. Run the Two-Phase Chain Tester$$

```bash
# Run the interactive chain tester
./chain-tester
```

### 3. **Two-Phase Workflow Explained**

#### **Phase 1: Setup & Proposal Creation**

1. **Chain Setup**: Automated initialization, key generation, genesis configuration
2. **Proposal Creation**: Interactive input for bridge workers and contract addresses
3. **Metadata Creation**: Creates `metadata.json` with proposal details
4. **IPFS Upload**: User uploads metadata to IPFS and provides CID
5. **Proposal JSON**: Creates `proposal.json` with IPFS metadata reference
6. **State Saving**: All inputs saved to `testing_state.json`
7. **Chain Startup**: Chain starts in background with 10-second countdown
8. **New Terminal**: Automatically opens new terminal for Phase 2

#### **Phase 2: Proposal Submission & Voting**

1. **State Loading**: Automatically loads all previous user inputs
2. **Proposal Submission**: Submits proposal to running chain
3. **Voting**: Interactive voting with multiple options
4. **Cleanup**: Automatically cleans up state files

### 4. **What You'll See**

```
🚀 Junction Chain Testing Script
=================================

📋 Cleaning up existing junctiond directory
✅ Cleaning up existing junctiond directory completed!

📋 Initializing junctiond node
✅ Initializing junctiond node completed!

📋 Generating keys
✅ Using existing key: test1

... (chain setup continues) ...

🔧 Creating Parameter Change Proposal
====================================

📝 Please enter bridge worker addresses (comma-separated):
Bridge Workers: air1h58eezgk5j4jwwpk3nxggx63gfuhnfcj78z5vj
Bridge Contract Address: 0x1234567890123456789012345678901234567890

✅ Proposal JSON created: proposal.json

🤔 Do you want to submit this proposal? (y/n): y

🚀 Starting chain in 10 seconds...
📋 Opening new terminal for proposal submission...
⏰ Starting chain in 10 seconds...
⏰ Starting chain in 9 seconds...
...
⏳ Waiting for chain to initialize...

# New terminal opens automatically
📤 Proposal Submission Phase
============================
⏳ Waiting for chain to be ready...
📋 Submitting parameter change proposal
✅ Submitting parameter change proposal completed!

🗳️ Do you want to vote on this proposal? (y/n): y
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

## 📁 **Generated Files**

The script creates several files during execution:

### **State Management**

- `testing_state.json` - **Persistent state file** containing all user inputs and phase information
- `metadata.json` - **Proposal metadata file** for IPFS upload
- `proposal.json` - Parameter change proposal file with IPFS metadata reference

### **Chain Data**

- `~/.junction/` - Chain data directory
- `~/.junction/config/genesis.json` - Modified genesis file with custom voting periods

### **State File Structure**

```json
{
  "phase": "proposal_submission",
  "bridge_workers": ["air1abc...", "air1def..."],
  "contract_address": "0x1234567890123456789012345678901234567890",
  "proposal_title": "Update EVM Bridge Authorized Unlockers",
  "proposal_description": "Add new addresses to the authorized unlockers list",
  "proposal_created": true,
  "chain_running": false,
  "proposal_submitted": false
}
```

## 🔧 **Commands Executed**

The script automatically executes these commands in sequence:

### **Phase 1: Chain Setup & Proposal Creation**

1. **Cleanup**: `rm -rf ~/.junction`
2. **Initialize**: `junctiond init junction-testing --default-denom uamf --chain-id junction`
3. **Generate Keys**: `junctiond keys show test1 --keyring-backend os || junctiond keys add test1 --keyring-backend os`
4. **Add Account**: `junctiond genesis add-genesis-account test1 100000000000uamf --keyring-backend os`
5. **Stake Validator**: `junctiond genesis gentx test1 10000000000uamf --keyring-backend os --gas-prices 0.0025uamf --chain-id junction`
6. **Collect Gentx**: `junctiond genesis collect-gentxs`
7. **Modify Genesis**: `jq` commands to update voting periods
8. **Start Chain**: `junctiond start --minimum-gas-prices 0.00025uamf` (in background)
9. **Open Terminal**: `gnome-terminal -- bash -c "cd $(pwd) && ./chain-tester; exec bash"`

### **Phase 2: Proposal Submission & Voting**

1. **Submit Proposal**: `junctiond tx gov submit-proposal proposal.json --from test1 --chain-id junction --fees 100uamf --keyring-backend os --gas auto --gas-adjustment 1.5`
2. **Vote**: `junctiond tx gov vote <proposal_id> yes --from test1 --keyring-backend os --chain-id junction --gas auto --gas-adjustment 1.5`
3. **Query Status**: `junctiond query gov proposals --output json`

## 🎯 **Two-Phase System Benefits**

### **Why Two Phases?**

- **Chain Requirement**: Proposals can only be submitted when the chain is running
- **User Experience**: Seamless workflow without manual intervention
- **State Persistence**: All user inputs preserved between phases
- **Automation**: No need to manually start chain or open new terminals

### **Phase Detection**

The script automatically detects which phase to run:

- **First run**: Detects no state file → runs Phase 1
- **Subsequent runs**: Detects `testing_state.json` → runs Phase 2
- **Completion**: Automatically cleans up state files

### **Terminal Handling**

- **Primary Terminal**: Runs Phase 1 (chain setup)
- **Secondary Terminal**: Automatically opens for Phase 2 (proposal submission)
- **Fallback Support**: Uses `xterm` if `gnome-terminal` not available

## 📋 **Requirements**

- Go 1.21 or later
- `junctiond` binary in `./build/junctiond`
- `jq` for JSON processing
- `gnome-terminal` or `xterm` for automatic terminal opening
- **IPFS service** for metadata upload (Pinata, Web3.Storage, or IPFS Desktop)
- Proper permissions to execute binaries

## 🌐 **IPFS Upload Services**

The script requires uploading metadata to IPFS. You can use any of these services:

### **Recommended Services**

- **Pinata**: https://pinata.cloud/ (Free tier available)
- **Web3.Storage**: https://web3.storage/ (Free tier available)
- **IPFS Desktop**: https://github.com/ipfs/ipfs-desktop (Local IPFS node)

### **Automated IPFS CID**

You can set the `IPFS_CID` environment variable to skip the manual upload step:

```bash
export IPFS_CID="QmYourMetadataHashHere"
./chain-tester
```

## 🔧 **Troubleshooting**

### **Common Issues**

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

4. **Terminal Not Opening**: Install required terminal emulator

   ```bash
   # Ubuntu/Debian
   sudo apt-get install gnome-terminal

   # Or install xterm as fallback
   sudo apt-get install xterm
   ```

### **Two-Phase Specific Issues**

5. **State File Issues**: Clear state and restart

   ```bash
   rm testing_state.json
   ./chain-tester
   ```

6. **Chain Not Starting**: Check if port is already in use

   ```bash
   # Check if chain is already running
   ps aux | grep junctiond

   # Kill existing process if needed
   pkill junctiond
   ```

7. **Cannot Stop Chain**: Use proper signal handling

   ```bash
   # The script now handles Ctrl+C properly
   # Press Ctrl+C to gracefully stop the chain

   # If that doesn't work, force kill:
   pkill -f junctiond
   ```

8. **Proposal Submission Fails**: Ensure chain is fully started

   ```bash
   # Wait longer for chain to initialize
   # The script automatically waits 15 seconds, but you may need more
   ```

9. **Environment Variables Not Loading**: Check `.env` file format
   ```bash
   # Ensure no spaces around =
   KEY=value
   # Not: KEY = value
   ```

### **Debug Mode**

9. **Enable Verbose Logging**: Check chain logs

   ```bash
   # Check junctiond logs
   tail -f ~/.junction/logs/junctiond.log
   ```

10. **Manual State Reset**: Clear all state files
    ```bash
    rm testing_state.json
    rm proposal.json
    rm -rf ~/.junction
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
