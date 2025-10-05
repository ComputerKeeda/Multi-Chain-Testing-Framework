package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var chainProcess *exec.Cmd

type ProposalMetadata struct {
	Title             string   `json:"title"`
	Authors           []string `json:"authors"`
	Summary           string   `json:"summary"`
	Details           string   `json:"details"`
	ProposalForumURL  string   `json:"proposal_forum_url"`
	VoteOptionContext string   `json:"vote_option_context"`
}

type ProposalMessage struct {
	Type      string `json:"@type"`
	Authority string `json:"authority"`
	Params    struct {
		BridgeWorkers         []string `json:"bridge_workers"`
		BridgeContractAddress string   `json:"bridge_contract_address"`
	} `json:"params"`
}

type Proposal struct {
	Messages  []ProposalMessage `json:"messages"`
	Metadata  string            `json:"metadata"`
	Deposit   string            `json:"deposit"`
	Title     string            `json:"title"`
	Summary   string            `json:"summary"`
	Expedited bool              `json:"expedited"`
}

type ChainConfig struct {
	Moniker          string
	ChainID          string
	Denom            string
	KeyName          string
	Amount           string
	ValidatorStake   string
	GasPrices        string
	MinimumGasPrices string
}

type TestingState struct {
	Phase             string   `json:"phase"`
	BridgeWorkers     []string `json:"bridge_workers"`
	ContractAddress   string   `json:"contract_address"`
	ProposalTitle     string   `json:"proposal_title"`
	ProposalSummary   string   `json:"proposal_summary"`
	ProposalDetails   string   `json:"proposal_details"`
	ProposalForumURL  string   `json:"proposal_forum_url"`
	IPFSCID           string   `json:"ipfs_cid"`
	ProposalCreated   bool     `json:"proposal_created"`
	ChainRunning      bool     `json:"chain_running"`
	ProposalSubmitted bool     `json:"proposal_submitted"`
}

func loadConfig() *ChainConfig {
	return &ChainConfig{
		Moniker:          getEnv("MONIKER", "junction-testing"),
		ChainID:          getEnv("CHAIN_ID", "junction"),
		Denom:            getEnv("DENOM", "uamf"),
		KeyName:          getEnv("KEY_NAME", "test1"),
		Amount:           getEnv("AMOUNT", "100000000000uamf"),
		ValidatorStake:   getEnv("VALIDATOR_STAKE", "10000000000uamf"),
		GasPrices:        getEnv("GAS_PRICES", "0.0025uamf"),
		MinimumGasPrices: getEnv("MINIMUM_GAS_PRICES", "0.00025uamf"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func loadEnvFile(filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not read .env file: %v\n", err)
		return
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}
}

func saveState(state *TestingState) error {
	stateJSON, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("testing_state.json", stateJSON, 0644)
}

func loadState() *TestingState {
	content, err := os.ReadFile("testing_state.json")
	if err != nil {
		return &TestingState{Phase: "setup"}
	}

	var state TestingState
	json.Unmarshal(content, &state)
	return &state
}

func clearState() {
	os.Remove("testing_state.json")
}

func setupSignalHandling() {
	// Create a channel to receive OS signals
	sigChan := make(chan os.Signal, 1)

	// Register the channel to receive SIGINT (Ctrl+C) and SIGTERM
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start a goroutine to handle signals
	go func() {
		sig := <-sigChan
		fmt.Printf("\n🛑 Received signal: %v\n", sig)
		fmt.Println("🔄 Stopping chain and cleaning up...")

		// Kill the chain process if it's running
		if chainProcess != nil && chainProcess.Process != nil {
			fmt.Println("⏹️  Stopping junctiond process...")
			// Try graceful shutdown first
			chainProcess.Process.Signal(syscall.SIGTERM)

			// Wait a bit for graceful shutdown
			time.Sleep(2 * time.Second)

			// Force kill if still running
			if chainProcess.ProcessState == nil || !chainProcess.ProcessState.Exited() {
				chainProcess.Process.Kill()
			}
		}

		// Also kill any other junctiond processes
		exec.Command("pkill", "junctiond").Run()

		// Clear state files
		clearState()

		fmt.Println("✅ Cleanup completed. Goodbye!")
		os.Exit(0)
	}()
}

func main() {
	fmt.Println("🚀 Junction Chain Testing Script")
	fmt.Println("=================================")

	// Set up signal handling for graceful shutdown
	setupSignalHandling()

	// Load configuration from environment variables
	config := loadConfig()

	// Check if .env file exists and load it
	if _, err := os.Stat(".env"); err == nil {
		loadEnvFile(".env")
	}

	// Load previous state if exists
	state := loadState()

	// Check if we're in proposal submission phase
	if state.Phase == "proposal_submission" {
		handleProposalSubmission(config, state)
		return
	}

	// Phase 1: Chain setup and proposal creation
	handleChainSetup(config, state)
}

func handleChainSetup(config *ChainConfig, state *TestingState) {
	// Step 1: Clean up existing directory
	executeStep("Cleaning up existing junctiond directory", func() error {
		return exec.Command("rm", "-rf", os.Getenv("HOME")+"/.junction").Run()
	})

	// Step 2: Initialize the junctiond node
	executeStep("Initializing junctiond node", func() error {
		cmd := exec.Command("./build/junctiond", "init", config.Moniker, "--default-denom", config.Denom, "--chain-id", config.ChainID)
		return cmd.Run()
	})

	// Step 3: Generate keys (or use existing)
	executeStep("Generating keys", func() error {
		// First check if key already exists
		checkCmd := exec.Command("./build/junctiond", "keys", "show", config.KeyName, "--keyring-backend", "os")
		err := checkCmd.Run()

		if err != nil {
			// Key doesn't exist, create it
			fmt.Printf("🔑 Creating new key: %s\n", config.KeyName)
			cmd := exec.Command("./build/junctiond", "keys", "add", config.KeyName, "--keyring-backend", "os")
			return cmd.Run()
		} else {
			// Key already exists, use it
			fmt.Printf("✅ Using existing key: %s\n", config.KeyName)
			return nil
		}
	})

	// Step 4: Add genesis account
	executeStep("Adding genesis account", func() error {
		cmd := exec.Command("./build/junctiond", "genesis", "add-genesis-account", config.KeyName, config.Amount, "--keyring-backend", "os")
		return cmd.Run()
	})

	// Step 5: Stake validator account
	executeStep("Staking validator account", func() error {
		cmd := exec.Command("./build/junctiond", "genesis", "gentx", config.KeyName, config.ValidatorStake, "--keyring-backend", "os", "--gas-prices", config.GasPrices, "--chain-id", config.ChainID)
		return cmd.Run()
	})

	// Step 6: Collect gentx files
	executeStep("Collecting gentx files", func() error {
		cmd := exec.Command("./build/junctiond", "genesis", "collect-gentxs")
		return cmd.Run()
	})

	// Step 7: Modify genesis file
	executeStep("Modifying genesis file with voting periods", func() error {
		genesisFile := os.Getenv("HOME") + "/.junction/config/genesis.json"
		cmd := exec.Command("jq",
			`.app_state.gov.params.max_deposit_period = "600s" |
			.app_state.gov.params.voting_period = "600s" |
			.app_state.gov.params.expedited_voting_period = "300s"`,
			genesisFile)

		output, err := cmd.Output()
		if err != nil {
			return err
		}

		return os.WriteFile(genesisFile+".tmp", output, 0644)
	})

	// Step 8: Move the modified genesis file
	executeStep("Applying genesis file changes", func() error {
		genesisFile := os.Getenv("HOME") + "/.junction/config/genesis.json"
		return exec.Command("mv", genesisFile+".tmp", genesisFile).Run()
	})

	// Step 9: Create parameter change proposal
	createParameterChangeProposal(config)

	// Step 10: Start the node
	fmt.Println("\n🎯 Starting junctiond node...")
	fmt.Println("Command: ./build/junctiond start --minimum-gas-prices", config.MinimumGasPrices)

	cmd := exec.Command("./build/junctiond", "start", "--minimum-gas-prices", config.MinimumGasPrices)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func executeStep(description string, action func() error) {
	fmt.Printf("\n📋 %s\n", description)
	fmt.Printf("Command: %s\n", getCommandDescription(description))

	// Show loading animation
	done := make(chan bool)
	go showLoadingAnimation(done)

	err := action()
	done <- true

	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ %s completed successfully!\n", description)
}

func showLoadingAnimation(done chan bool) {
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	for {
		select {
		case <-done:
			fmt.Print("\r")
			return
		default:
			fmt.Printf("\r%s Processing...", spinner[i%len(spinner)])
			time.Sleep(100 * time.Millisecond)
			i++
		}
	}
}

func getCommandDescription(description string) string {
	descriptions := map[string]string{
		"Cleaning up existing junctiond directory":   "rm -rf ~/.junction",
		"Initializing junctiond node":                "junctiond init junction-testing --default-denom uamf --chain-id junction",
		"Generating keys":                            "junctiond keys show test1 --keyring-backend os || junctiond keys add test1 --keyring-backend os",
		"Adding genesis account":                     "junctiond genesis add-genesis-account test1 100000000000uamf --keyring-backend os",
		"Staking validator account":                  "junctiond genesis gentx test1 10000000000uamf --keyring-backend os --gas-prices 0.0025uamf --chain-id junction",
		"Collecting gentx files":                     "junctiond genesis collect-gentxs",
		"Modifying genesis file with voting periods": "jq command to update voting periods",
		"Applying genesis file changes":              "mv genesis.json.tmp genesis.json",
	}
	return descriptions[description]
}

func createParameterChangeProposal(config *ChainConfig) {
	fmt.Println("\n🔧 Creating Parameter Change Proposal")
	fmt.Println("====================================")

	// Check for environment variables first
	envWorkers := getEnv("BRIDGE_WORKERS", "")
	envContract := getEnv("BRIDGE_CONTRACT_ADDRESS", "")

	var bridgeWorkers []string
	var contractAddress string

	// Use environment variables if available
	if envWorkers != "" {
		workers := strings.Split(envWorkers, ",")
		for _, worker := range workers {
			bridgeWorkers = append(bridgeWorkers, strings.TrimSpace(worker))
		}
		fmt.Printf("✅ Using bridge workers from environment: %v\n", bridgeWorkers)
	} else {
		// Collect bridge workers interactively
		fmt.Println("\n📝 Please enter bridge worker addresses (comma-separated):")
		fmt.Print("Bridge Workers: ")
		reader := bufio.NewReader(os.Stdin)
		workersInput, _ := reader.ReadString('\n')
		workersInput = strings.TrimSpace(workersInput)

		if workersInput != "" {
			workers := strings.Split(workersInput, ",")
			for _, worker := range workers {
				bridgeWorkers = append(bridgeWorkers, strings.TrimSpace(worker))
			}
		} else {
			// Default values for testing
			bridgeWorkers = []string{"air1abc...", "air1def...", "air1ghi..."}
			fmt.Println("Using default addresses for testing")
		}
	}

	// Use environment variable if available
	if envContract != "" {
		contractAddress = envContract
		fmt.Printf("✅ Using contract address from environment: %s\n", contractAddress)
	} else {
		// Collect bridge contract address interactively
		fmt.Print("Bridge Contract Address: ")
		reader := bufio.NewReader(os.Stdin)
		contractInput, _ := reader.ReadString('\n')
		contractAddress = strings.TrimSpace(contractInput)

		if contractAddress == "" {
			contractAddress = "0x1234567890123456789012345678901234567890"
			fmt.Println("Using default contract address for testing")
		}
	}

	// Collect additional proposal information
	fmt.Println("\n📋 Additional Proposal Information")
	fmt.Println("==================================")

	// Get proposal title
	proposalTitle := getEnv("PROPOSAL_TITLE", "Update EVM Bridge Authorized Unlockers")
	fmt.Printf("Proposal Title [%s]: ", proposalTitle)
	reader := bufio.NewReader(os.Stdin)
	titleInput, _ := reader.ReadString('\n')
	titleInput = strings.TrimSpace(titleInput)
	if titleInput != "" {
		proposalTitle = titleInput
	}

	// Get proposal summary
	proposalSummary := getEnv("PROPOSAL_SUMMARY", "This proposal aims to update the EVM bridge authorized unlockers list and add new bridge contract addresses to enhance the bridge's security and functionality.")
	fmt.Printf("Proposal Summary [%s]: ", proposalSummary)
	summaryInput, _ := reader.ReadString('\n')
	summaryInput = strings.TrimSpace(summaryInput)
	if summaryInput != "" {
		proposalSummary = summaryInput
	}

	// Get proposal details
	proposalDetails := getEnv("PROPOSAL_DETAILS", "The EVM bridge requires regular updates to its authorized unlockers list to maintain security and add new trusted validators. This proposal adds the following addresses to the authorized unlockers list and updates the bridge contract address to ensure proper bridge operations.")
	fmt.Printf("Proposal Details [%s]: ", proposalDetails)
	detailsInput, _ := reader.ReadString('\n')
	detailsInput = strings.TrimSpace(detailsInput)
	if detailsInput != "" {
		proposalDetails = detailsInput
	}

	// Get proposal forum URL
	proposalForumURL := getEnv("PROPOSAL_FORUM_URL", "https://forum.junction.network/t/update-evm-bridge-authorized-unlockers")
	fmt.Printf("Proposal Forum URL [%s]: ", proposalForumURL)
	forumInput, _ := reader.ReadString('\n')
	forumInput = strings.TrimSpace(forumInput)
	if forumInput != "" {
		proposalForumURL = forumInput
	}

	// Create metadata JSON
	metadata := ProposalMetadata{
		Title:             proposalTitle,
		Authors:           []string{config.KeyName},
		Summary:           proposalSummary,
		Details:           proposalDetails,
		ProposalForumURL:  proposalForumURL,
		VoteOptionContext: "yes,no,abstain",
	}

	// Write metadata to JSON file
	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		fmt.Printf("❌ Error creating metadata JSON: %v\n", err)
		return
	}

	err = os.WriteFile("metadata.json", metadataJSON, 0644)
	if err != nil {
		fmt.Printf("❌ Error writing metadata file: %v\n", err)
		return
	}

	fmt.Println("✅ Metadata JSON created: metadata.json")

	// Check for IPFS CID in environment variables first
	envCID := getEnv("IPFS_CID", "")
	var cidInput string

	if envCID != "" {
		cidInput = envCID
		fmt.Printf("✅ Using IPFS CID from environment: %s\n", cidInput)
	} else {
		// Wait for user to upload metadata to IPFS and get CID
		fmt.Println("\n📤 IPFS Upload Step")
		fmt.Println("===================")
		fmt.Println("Please upload the metadata.json file to IPFS and get the CID.")
		fmt.Println("You can use services like:")
		fmt.Println("  - Pinata: https://pinata.cloud/")
		fmt.Println("  - IPFS Desktop: https://github.com/ipfs/ipfs-desktop")
		fmt.Println("  - Web3.Storage: https://web3.storage/")
		fmt.Println("  - Or any other IPFS service")
		fmt.Println("")
		fmt.Print("Enter the IPFS CID (e.g., QmYourHashHere): ")
		reader := bufio.NewReader(os.Stdin)
		cidInput, _ = reader.ReadString('\n')
		cidInput = strings.TrimSpace(cidInput)

		if cidInput == "" {
			fmt.Println("❌ CID is required to continue. Please upload the metadata and get the CID.")
			return
		}

		// Validate CID format (basic check)
		if !strings.HasPrefix(cidInput, "Qm") && !strings.HasPrefix(cidInput, "bafy") {
			fmt.Printf("⚠️  Warning: CID doesn't look like a standard IPFS hash. Continuing anyway...\n")
		}

		fmt.Printf("✅ Using IPFS CID: %s\n", cidInput)
	}

	// Create proposal JSON with the actual IPFS CID
	proposal := Proposal{
		Messages: []ProposalMessage{
			{
				Type:      "/junction.evmbridge.MsgUpdateParams",
				Authority: "air10d07y265gmmuvt4z0w9aw880jnsr700jszsute",
				Params: struct {
					BridgeWorkers         []string `json:"bridge_workers"`
					BridgeContractAddress string   `json:"bridge_contract_address"`
				}{
					BridgeWorkers:         bridgeWorkers,
					BridgeContractAddress: contractAddress,
				},
			},
		},
		Metadata:  "ipfs://" + cidInput,
		Deposit:   "1000000uamf",
		Title:     proposalTitle,
		Summary:   proposalSummary,
		Expedited: true,
	}

	// Save state with proposal data
	state := loadState()
	state.BridgeWorkers = bridgeWorkers
	state.ContractAddress = contractAddress
	state.ProposalTitle = proposalTitle
	state.ProposalSummary = proposalSummary
	state.ProposalDetails = proposalDetails
	state.ProposalForumURL = proposalForumURL
	state.IPFSCID = cidInput
	saveState(state)

	// Write proposal to JSON file
	proposalJSON, err := json.MarshalIndent(proposal, "", "  ")
	if err != nil {
		fmt.Printf("❌ Error creating proposal JSON: %v\n", err)
		return
	}

	err = os.WriteFile("proposal.json", proposalJSON, 0644)
	if err != nil {
		fmt.Printf("❌ Error writing proposal file: %v\n", err)
		return
	}

	fmt.Println("✅ Proposal JSON created: proposal.json")

	// Ask if user wants to submit the proposal
	fmt.Print("\n🤔 Do you want to submit this proposal? (y/n): ")
	submitInput, _ := reader.ReadString('\n')
	submitInput = strings.TrimSpace(strings.ToLower(submitInput))

	if submitInput == "y" || submitInput == "yes" {
		// Save state for proposal submission phase
		state.Phase = "proposal_submission"
		state.ProposalCreated = true
		saveState(state)

		fmt.Println("\n🚀 Starting chain in 10 seconds...")
		fmt.Println("📋 Opening new terminal for proposal submission...")

		// Countdown
		for i := 10; i > 0; i-- {
			fmt.Printf("\r⏰ Starting chain in %d seconds...", i)
			time.Sleep(1 * time.Second)
		}
		fmt.Println()

		// Start chain in background
		go startChain(config)

		// Wait a bit for chain to start
		fmt.Println("⏳ Waiting for chain to initialize...")
		time.Sleep(15 * time.Second)

		// Open new terminal for proposal submission
		openNewTerminal()
	} else {
		// Start chain normally
		startChain(config)
	}
}

func startChain(config *ChainConfig) {
	// Check if chain is already running
	if isChainRunning() {
		fmt.Println("⚠️  Junctiond is already running!")
		fmt.Println("💡 If you want to restart, please stop the existing process first")
		fmt.Println("   You can use: pkill junctiond")
		return
	}

	fmt.Println("🚀 Starting junctiond node...")
	cmd := exec.Command("./build/junctiond", "start", "--minimum-gas-prices", config.MinimumGasPrices)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Store the process reference for signal handling
	chainProcess = cmd

	// Start the process
	err := cmd.Start()
	if err != nil {
		fmt.Printf("❌ Error starting junctiond: %v\n", err)
		return
	}

	fmt.Println("✅ Junctiond started successfully!")
	fmt.Println("💡 Press Ctrl+C to stop the chain and exit")

	// Wait for the process to complete
	cmd.Wait()
}

func isChainRunning() bool {
	// Check if junctiond process is running
	cmd := exec.Command("pgrep", "junctiond")
	err := cmd.Run()
	return err == nil
}

func openNewTerminal() {
	// Open new terminal and run proposal submission
	cmd := exec.Command("gnome-terminal", "--", "bash", "-c", "cd $(pwd) && ./chain-tester; exec bash")
	if err := cmd.Run(); err != nil {
		// Fallback for other terminals
		exec.Command("xterm", "-e", "cd $(pwd) && ./chain-tester").Run()
	}
}

func handleProposalSubmission(config *ChainConfig, state *TestingState) {
	fmt.Println("📤 Proposal Submission Phase")
	fmt.Println("============================")

	if !state.ProposalCreated {
		fmt.Println("❌ No proposal found. Please run the setup phase first.")
		return
	}

	// Wait a bit more for chain to be ready
	fmt.Println("⏳ Waiting for chain to be ready...")
	time.Sleep(10 * time.Second)

	// Submit the proposal
	submitProposal()

	// Ask if user wants to vote
	fmt.Print("\n🗳️  Do you want to vote on this proposal? (y/n): ")
	reader := bufio.NewReader(os.Stdin)
	voteInput, _ := reader.ReadString('\n')
	voteInput = strings.TrimSpace(strings.ToLower(voteInput))

	if voteInput == "y" || voteInput == "yes" {
		voteOnProposal()
	}

	// Clear state after completion
	clearState()
	fmt.Println("✅ Testing completed!")
}

func submitProposal() {
	fmt.Println("\n📤 Submitting Parameter Change Proposal")
	fmt.Println("======================================")

	proposerKey := getEnv("PROPOSER_KEY", "test1")
	chainID := getEnv("CHAIN_ID", "junction")
	fees := getEnv("PROPOSAL_FEES", "100uamf")

	// Show the command that will be executed
	fmt.Printf("Command: junctiond tx gov submit-proposal proposal.json --from %s --chain-id %s --fees %s\n", proposerKey, chainID, fees)

	// Execute the proposal submission
	executeStep("Submitting parameter change proposal", func() error {
		cmd := exec.Command("./build/junctiond", "tx", "gov", "submit-proposal", "proposal.json", "--from", proposerKey, "--chain-id", chainID, "--fees", fees, "--keyring-backend", "os", "--gas", "auto", "--gas-adjustment", "1.5")
		return cmd.Run()
	})

	// Ask if user wants to vote on the proposal
	fmt.Print("\n🗳️  Do you want to vote on this proposal? (y/n): ")
	reader := bufio.NewReader(os.Stdin)
	voteInput, _ := reader.ReadString('\n')
	voteInput = strings.TrimSpace(strings.ToLower(voteInput))

	if voteInput == "y" || voteInput == "yes" {
		voteOnProposal()
	}
}

func voteOnProposal() {
	fmt.Println("\n🗳️  Voting on Proposal")
	fmt.Println("=====================")

	// Get proposal ID
	fmt.Print("Enter Proposal ID: ")
	reader := bufio.NewReader(os.Stdin)
	proposalIDInput, _ := reader.ReadString('\n')
	proposalID := strings.TrimSpace(proposalIDInput)

	if proposalID == "" {
		fmt.Println("❌ Proposal ID is required")
		return
	}

	// Check for environment variable vote option
	envVote := getEnv("VOTE_OPTION", "")
	var vote string

	if envVote != "" {
		vote = envVote
		fmt.Printf("✅ Using vote option from environment: %s\n", vote)
	} else {
		// Get vote option interactively
		fmt.Println("Vote options:")
		fmt.Println("1. yes")
		fmt.Println("2. no")
		fmt.Println("3. no_with_veto")
		fmt.Println("4. abstain")
		fmt.Print("Enter vote option (1-4): ")

		voteOptionInput, _ := reader.ReadString('\n')
		voteOption := strings.TrimSpace(voteOptionInput)

		switch voteOption {
		case "1":
			vote = "yes"
		case "2":
			vote = "no"
		case "3":
			vote = "no_with_veto"
		case "4":
			vote = "abstain"
		default:
			fmt.Println("❌ Invalid vote option")
			return
		}
	}

	proposerKey := getEnv("PROPOSER_KEY", "test1")
	chainID := getEnv("CHAIN_ID", "junction")

	// Execute vote
	executeStep("Voting on proposal", func() error {
		cmd := exec.Command("./build/junctiond", "tx", "gov", "vote", proposalID, vote, "--from", proposerKey, "--keyring-backend", "os", "--chain-id", chainID, "--gas", "auto", "--gas-adjustment", "1.5")
		return cmd.Run()
	})

	// Ask if user wants to wait for voting period
	fmt.Print("\n⏰ Do you want to wait for the voting period to complete? (y/n): ")
	waitInput, _ := reader.ReadString('\n')
	waitInput = strings.TrimSpace(strings.ToLower(waitInput))

	if waitInput == "y" || waitInput == "yes" {
		waitForVotingPeriod()
	}
}

func waitForVotingPeriod() {
	fmt.Println("\n⏰ Waiting for Voting Period to Complete")
	fmt.Println("=====================================")

	// Check for environment variable first
	envDuration := getEnv("VOTING_PERIOD", "")
	var duration int

	if envDuration != "" {
		if d, err := strconv.Atoi(envDuration); err == nil {
			duration = d
			fmt.Printf("✅ Using voting period from environment: %d seconds\n", duration)
		} else {
			duration = 600
			fmt.Printf("⚠️  Invalid VOTING_PERIOD in environment, using default: %d seconds\n", duration)
		}
	} else {
		// Get voting period duration interactively
		fmt.Print("Enter voting period duration in seconds (default: 600): ")
		reader := bufio.NewReader(os.Stdin)
		durationInput, _ := reader.ReadString('\n')
		durationInput = strings.TrimSpace(durationInput)

		duration = 600 // default 10 minutes
		if durationInput != "" {
			if d, err := strconv.Atoi(durationInput); err == nil {
				duration = d
			}
		}
	}

	fmt.Printf("⏳ Waiting for %d seconds...\n", duration)

	// Show countdown animation
	done := make(chan bool)
	go showCountdownAnimation(duration, done)

	time.Sleep(time.Duration(duration) * time.Second)
	done <- true

	fmt.Println("\n✅ Voting period completed!")

	// Query proposal status
	executeStep("Querying proposal status", func() error {
		cmd := exec.Command("./build/junctiond", "query", "gov", "proposals", "--output", "json")
		output, err := cmd.Output()
		if err != nil {
			return err
		}
		fmt.Printf("📊 Proposal Status:\n%s\n", string(output))
		return nil
	})
}

func showCountdownAnimation(duration int, done chan bool) {
	for i := duration; i > 0; i-- {
		select {
		case <-done:
			return
		default:
			minutes := i / 60
			seconds := i % 60
			fmt.Printf("\r⏰ Time remaining: %02d:%02d", minutes, seconds)
			time.Sleep(1 * time.Second)
		}
	}
	fmt.Print("\r⏰ Time remaining: 00:00")
}
