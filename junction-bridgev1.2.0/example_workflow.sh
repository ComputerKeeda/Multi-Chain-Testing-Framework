#!/bin/bash

# Junction Bridge Testing Workflow Example
# This script demonstrates the complete workflow for testing governance proposals

echo "🚀 Junction Bridge Testing Workflow"
echo "=================================="

# Step 1: Build the tool
echo "📦 Building the junction-bridge tool..."
chmod +x build_executable.sh
./build_executable.sh

# Step 2: Initialize the node (this will start the blockchain)
echo "🔧 Initializing Junction node..."
echo "Note: This will start the blockchain node. Press Ctrl+C to stop when ready to proceed."
echo "Press Enter to continue..."
read

# Step 3: Submit a governance proposal
echo "🗳️  Submitting governance proposal..."
echo "This will:"
echo "1. Create metadata.json from draft template"
echo "2. Guide you to upload it to IPFS"
echo "3. Create proposal.json with the IPFS CID"
echo "4. Submit the proposal to the chain"
echo "Press Enter to continue..."
read
./build/junction-bridge submit-proposal

# Step 4: Wait for deposit period to end
echo "⏳ Waiting for deposit period to end..."
echo "The proposal is now in deposit period. Wait for it to enter voting period."
echo "Press Enter when ready to vote..."
read

# Step 5: Vote on the proposal
echo "🗳️  Voting on the proposal..."
echo "Enter the proposal ID (you can get this from the monitor or REST API):"
read -p "Proposal ID: " proposal_id
echo "Vote options: yes, no, abstain, no_with_veto"
read -p "Your vote: " vote_option
./build/junction-bridge vote $proposal_id $vote_option

# Step 6: Monitor the proposal
echo "🔍 Monitoring proposal status..."
echo "This will show real-time status with animations."
echo "Press Ctrl+C to stop monitoring when done."
echo "Press Enter to start monitoring..."
read
./build/junction-bridge monitor-proposals

echo "✅ Workflow completed!"
echo "The proposal has been submitted, voted on, and monitored."
