package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type ProposalChange struct {
	Subspace string      `json:"subspace"`
	Key      string      `json:"key"`
	Value    interface{} `json:"value"`
}

type Proposal struct {
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Changes     []ProposalChange `json:"changes"`
}

type BridgeParams struct {
	BridgeWorkers         []string `json:"bridge_workers"`
	BridgeContractAddress string   `json:"bridge_contract_address"`
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

func main() {
	fmt.Println("🚀 Junction Chain Testing Script")
	fmt.Println("=================================")

	// Load configuration from environment variables
	config := loadConfig()

	// Check if .env file exists and load it
	if _, err := os.Stat(".env"); err == nil {
		loadEnvFile(".env")
	}

	// Step 1: Clean up existing directory
	executeStep("Cleaning up existing junctiond directory", func() error {
		return exec.Command("rm", "-rf", os.Getenv("HOME")+"/.junction").Run()
	})

	// Step 2: Initialize the junctiond node
	executeStep("Initializing junctiond node", func() error {
		cmd := exec.Command("./build/junctiond", "init", config.Moniker, "--default-denom", config.Denom, "--chain-id", config.ChainID)
		return cmd.Run()
	})

	// Step 3: Generate keys
	executeStep("Generating keys", func() error {
		cmd := exec.Command("./build/junctiond", "keys", "add", config.KeyName, "--keyring-backend", "os")
		return cmd.Run()
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
	createParameterChangeProposal()

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
		"Generating keys":                            "junctiond keys add test1 --keyring-backend os",
		"Adding genesis account":                     "junctiond genesis add-genesis-account test1 100000000000uamf --keyring-backend os",
		"Staking validator account":                  "junctiond genesis gentx test1 10000000000uamf --keyring-backend os --gas-prices 0.0025uamf --chain-id junction",
		"Collecting gentx files":                     "junctiond genesis collect-gentxs",
		"Modifying genesis file with voting periods": "jq command to update voting periods",
		"Applying genesis file changes":              "mv genesis.json.tmp genesis.json",
	}
	return descriptions[description]
}

func createParameterChangeProposal() {
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

	// Create proposal
	proposal := Proposal{
		Title:       getEnv("PROPOSAL_TITLE", "Update EVM Bridge Authorized Unlockers"),
		Description: getEnv("PROPOSAL_DESCRIPTION", "Add new addresses to the authorized unlockers list"),
		Changes: []ProposalChange{
			{
				Subspace: "evmbridge",
				Key:      "params",
				Value: BridgeParams{
					BridgeWorkers:         bridgeWorkers,
					BridgeContractAddress: contractAddress,
				},
			},
		},
	}

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
	reader := bufio.NewReader(os.Stdin)
	submitInput, _ := reader.ReadString('\n')
	submitInput = strings.TrimSpace(strings.ToLower(submitInput))

	if submitInput == "y" || submitInput == "yes" {
		submitProposal()
	}
}

func submitProposal() {
	fmt.Println("\n📤 Submitting Parameter Change Proposal")
	fmt.Println("======================================")

	proposerKey := getEnv("PROPOSER_KEY", "test1")
	deposit := getEnv("PROPOSAL_DEPOSIT", "1000000uamf")
	chainID := getEnv("CHAIN_ID", "junction")

	// Show the command that will be executed
	fmt.Printf("Command: junctiond tx gov submit-proposal param-change proposal.json --from %s --deposit %s\n", proposerKey, deposit)

	// Execute the proposal submission
	executeStep("Submitting parameter change proposal", func() error {
		cmd := exec.Command("./build/junctiond", "tx", "gov", "submit-proposal", "param-change", "proposal.json", "--from", proposerKey, "--deposit", deposit, "--keyring-backend", "os", "--chain-id", chainID, "--gas", "auto", "--gas-adjustment", "1.5")
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
