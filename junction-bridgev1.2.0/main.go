package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	Moniker          string `mapstructure:"moniker"`
	ChainID          string `mapstructure:"chain_id"`
	Denom            string `mapstructure:"denom"`
	KeyName          string `mapstructure:"key_name"`
	Amount           string `mapstructure:"amount"`
	ValidatorStake   string `mapstructure:"validator_stake"`
	JunctiondPath    string `mapstructure:"junctiond_path"`
	HomeDir          string `mapstructure:"home_dir"`
	MinimumGasPrices string `mapstructure:"minimum_gas_prices"`
	RestEndpoint     string `mapstructure:"rest_endpoint"`
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

type ProposalResponse struct {
	Proposals []struct {
		ID               string `json:"id"`
		Status           string `json:"status"`
		VotingStartTime  string `json:"voting_start_time"`
		VotingEndTime    string `json:"voting_end_time"`
		FinalTallyResult struct {
			YesCount        string `json:"yes_count"`
			AbstainCount    string `json:"abstain_count"`
			NoCount         string `json:"no_count"`
			NoWithVetoCount string `json:"no_with_veto_count"`
		} `json:"final_tally_result"`
	} `json:"proposals"`
}

type GenesisConfig struct {
	AppState struct {
		Gov struct {
			Params struct {
				MaxDepositPeriod      string `json:"max_deposit_period"`
				VotingPeriod          string `json:"voting_period"`
				ExpeditedVotingPeriod string `json:"expedited_voting_period"`
			} `json:"params"`
		} `json:"gov"`
	} `json:"app_state"`
}

var config Config

var rootCmd = &cobra.Command{
	Use:   "junction-bridge",
	Short: "Junction Bridge Testing Tool",
	Long:  "A tool for setting up and managing Junction blockchain nodes for bridge testing",
}

var initCmd = &cobra.Command{
	Use:   "init-node",
	Short: "Initialize and start a Junction node",
	Long:  "Initialize a Junction blockchain node with custom configuration and start it",
	Run:   runInitNode,
}

var submitProposalCmd = &cobra.Command{
	Use:   "submit-proposal",
	Short: "Submit a governance proposal",
	Long:  "Submit a governance proposal to update EVM bridge parameters",
	Run:   runSubmitProposal,
}

var voteCmd = &cobra.Command{
	Use:   "vote [proposal-id] [vote-option]",
	Short: "Vote on a governance proposal",
	Long:  "Vote on a governance proposal (yes/no/abstain/no_with_veto)",
	Args:  cobra.ExactArgs(2),
	Run:   runVote,
}

var monitorCmd = &cobra.Command{
	Use:   "monitor-proposals",
	Short: "Monitor proposal status",
	Long:  "Monitor the status of governance proposals with animations",
	Run:   runMonitorProposals,
}

func init() {
	// Initialize Viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.junction-bridge")

	// Set default values
	viper.SetDefault("moniker", "junction-testing")
	viper.SetDefault("chain_id", "junction")
	viper.SetDefault("denom", "uamf")
	viper.SetDefault("key_name", "test1")
	viper.SetDefault("amount", "100000000000uamf")
	viper.SetDefault("validator_stake", "10000000000uamf")
	viper.SetDefault("junctiond_path", "./build/junctiond")
	viper.SetDefault("home_dir", "$HOME/.junction")
	viper.SetDefault("minimum_gas_prices", "0.00025uamf")
	viper.SetDefault("rest_endpoint", "http://localhost:1317")

	// Bind flags
	initCmd.Flags().String("moniker", "junction-testing", "Moniker for the node")
	initCmd.Flags().String("chain-id", "junction", "Chain ID")
	initCmd.Flags().String("denom", "uamf", "Denomination")
	initCmd.Flags().String("key-name", "test1", "Key name")
	initCmd.Flags().String("amount", "100000000000uamf", "Initial amount")
	initCmd.Flags().String("validator-stake", "10000000000uamf", "Validator stake amount")
	initCmd.Flags().String("junctiond-path", "./build/junctiond", "Path to junctiond binary")
	initCmd.Flags().String("home-dir", "$HOME/.junction", "Home directory")
	initCmd.Flags().String("minimum-gas-prices", "0.00025uamf", "Minimum gas prices")

	viper.BindPFlags(initCmd.Flags())

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(submitProposalCmd)
	rootCmd.AddCommand(voteCmd)
	rootCmd.AddCommand(monitorCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runInitNode(cmd *cobra.Command, args []string) {
	// Load configuration
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Error reading config file: %v\n", err)
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("Error unmarshaling config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🚀 Starting Junction Node Initialization...")
	fmt.Printf("Moniker: %s\n", config.Moniker)
	fmt.Printf("Chain ID: %s\n", config.ChainID)
	fmt.Printf("Denom: %s\n", config.Denom)

	// Step 1: Remove existing junctiond directory
	fmt.Println("\n📁 Removing existing junctiond directory...")
	homeDir := os.ExpandEnv(config.HomeDir)
	if err := os.RemoveAll(homeDir); err != nil {
		fmt.Printf("Warning: Could not remove existing directory: %v\n", err)
	}

	// Step 2: Initialize the junctiond node
	fmt.Println("\n🔧 Initializing junctiond node...")
	initCmd := exec.Command(config.JunctiondPath, "init", config.Moniker, "--default-denom", config.Denom, "--chain-id", config.ChainID)
	if err := runCommand(initCmd); err != nil {
		fmt.Printf("Error initializing node: %v\n", err)
		os.Exit(1)
	}

	// Step 3: Generate keys (or use existing)
	fmt.Println("\n🔑 Generating keys...")

	// First check if key already exists
	checkKeyCmd := exec.Command(config.JunctiondPath, "keys", "show", config.KeyName, "--keyring-backend", "os")
	err := checkKeyCmd.Run()

	if err != nil {
		// Key doesn't exist, create it
		fmt.Printf("🔑 Creating new key: %s\n", config.KeyName)
		keyCmd := exec.Command(config.JunctiondPath, "keys", "add", config.KeyName, "--keyring-backend", "os")
		if err := runCommand(keyCmd); err != nil {
			fmt.Printf("Error generating keys: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Key already exists, use it
		fmt.Printf("✅ Using existing key: %s\n", config.KeyName)
	}

	// Step 4: Add genesis account
	fmt.Println("\n💰 Adding genesis account...")
	genesisAccountCmd := exec.Command(config.JunctiondPath, "genesis", "add-genesis-account", config.KeyName, config.Amount, "--keyring-backend", "os")
	if err := runCommand(genesisAccountCmd); err != nil {
		fmt.Printf("Error adding genesis account: %v\n", err)
		os.Exit(1)
	}

	// Step 5: Stake validator account
	fmt.Println("\n🏛️ Staking validator account...")
	gentxCmd := exec.Command(config.JunctiondPath, "genesis", "gentx", config.KeyName, config.ValidatorStake, "--keyring-backend", "os", "--gas-prices", "0.0025uamf", "--chain-id", config.ChainID)
	if err := runCommand(gentxCmd); err != nil {
		fmt.Printf("Error creating gentx: %v\n", err)
		os.Exit(1)
	}

	// Step 6: Collect gentx files
	fmt.Println("\n📋 Collecting gentx files...")
	collectGentxCmd := exec.Command(config.JunctiondPath, "genesis", "collect-gentxs")
	if err := runCommand(collectGentxCmd); err != nil {
		fmt.Printf("Error collecting gentx files: %v\n", err)
		os.Exit(1)
	}

	// Step 7: Modify genesis file
	fmt.Println("\n⚙️ Modifying genesis file...")
	if err := modifyGenesisFile(homeDir); err != nil {
		fmt.Printf("Error modifying genesis file: %v\n", err)
		os.Exit(1)
	}

	// Step 8: Modify app.toml file
	fmt.Println("\n🔧 Modifying app.toml file...")
	if err := modifyAppTomlFile(homeDir); err != nil {
		fmt.Printf("Error modifying app.toml file: %v\n", err)
		os.Exit(1)
	}

	// Step 9: Start the node
	fmt.Println("\n🚀 Starting junctiond node...")
	fmt.Println("Node will start with minimum gas prices:", config.MinimumGasPrices)

	startCmd := exec.Command(config.JunctiondPath, "start", "--minimum-gas-prices", config.MinimumGasPrices)
	startCmd.Stdout = os.Stdout
	startCmd.Stderr = os.Stderr

	if err := startCmd.Run(); err != nil {
		fmt.Printf("Error starting node: %v\n", err)
		os.Exit(1)
	}
}

func runCommand(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func modifyGenesisFile(homeDir string) error {
	genesisFile := filepath.Join(homeDir, "config", "genesis.json")

	// Read the genesis file
	data, err := os.ReadFile(genesisFile)
	if err != nil {
		return fmt.Errorf("error reading genesis file: %v", err)
	}

	// Parse JSON
	var genesis map[string]interface{}
	if err := json.Unmarshal(data, &genesis); err != nil {
		return fmt.Errorf("error parsing genesis file: %v", err)
	}

	// Navigate to app_state.gov.params and update values
	appState, ok := genesis["app_state"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("app_state not found in genesis file")
	}

	gov, ok := appState["gov"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("gov not found in app_state")
	}

	params, ok := gov["params"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("params not found in gov")
	}

	// Update the parameters
	params["max_deposit_period"] = "600s"
	params["voting_period"] = "660s"
	params["expedited_voting_period"] = "300s"

	// Write back to file
	updatedData, err := json.MarshalIndent(genesis, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling updated genesis: %v", err)
	}

	if err := os.WriteFile(genesisFile, updatedData, 0644); err != nil {
		return fmt.Errorf("error writing updated genesis file: %v", err)
	}

	fmt.Println("✅ Genesis file updated with new voting and deposit periods")
	return nil
}

func modifyAppTomlFile(homeDir string) error {
	appTomlFile := filepath.Join(homeDir, "config", "app.toml")

	// Read the app.toml file
	data, err := os.ReadFile(appTomlFile)
	if err != nil {
		return fmt.Errorf("error reading app.toml file: %v", err)
	}

	content := string(data)

	// Apply modifications
	content = strings.ReplaceAll(content, `minimum-gas-prices = ""`, `minimum-gas-prices = "0.00025uamf"`)
	content = strings.ReplaceAll(content, `enable = false`, `enable = true`)
	content = strings.ReplaceAll(content, `swagger = false`, `swagger = true`)

	// Write back to file
	if err := os.WriteFile(appTomlFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("error writing updated app.toml file: %v", err)
	}

	fmt.Println("✅ App.toml file updated with new minimum gas prices")
	return nil
}

func runSubmitProposal(cmd *cobra.Command, args []string) {
	// Load configuration
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Error reading config file: %v\n", err)
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("Error unmarshaling config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🗳️  Starting Governance Proposal Submission...")

	// Step 1: Create metadata.json from draft template
	fmt.Println("\n📝 Creating metadata.json from draft template...")

	// Read the draft metadata template
	draftMetadata, err := os.ReadFile("draft_metadata.json")
	if err != nil {
		fmt.Printf("Error reading draft_metadata.json: %v\n", err)
		fmt.Println("Please ensure draft_metadata.json exists in the current directory")
		os.Exit(1)
	}

	// Write metadata.json
	if err := os.WriteFile("metadata.json", draftMetadata, 0644); err != nil {
		fmt.Printf("Error creating metadata.json: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ metadata.json created successfully")
	fmt.Println("\n📤 Next steps:")
	fmt.Println("1. Upload metadata.json to IPFS")
	fmt.Println("2. Copy the IPFS CID (hash)")
	fmt.Println("3. Paste the CID below")
	fmt.Println("\nExample IPFS upload commands:")
	fmt.Println("  # Using ipfs CLI:")
	fmt.Println("  ipfs add metadata.json")
	fmt.Println("  # Or using web interface at https://ipfs.io/")
	fmt.Println("")
	fmt.Print("Enter IPFS CID: ")
	reader := bufio.NewReader(os.Stdin)
	ipfsCID, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		os.Exit(1)
	}
	ipfsCID = strings.TrimSpace(ipfsCID)

	// Step 2: Create proposal.json
	fmt.Println("\n📝 Creating proposal.json...")
	proposal := Proposal{
		Messages: []ProposalMessage{
			{
				Type:      "/junction.evmbridge.MsgUpdateParams",
				Authority: "air10d07y265gmmuvt4z0w9aw880jnsr700jszsute",
				Params: struct {
					BridgeWorkers         []string `json:"bridge_workers"`
					BridgeContractAddress string   `json:"bridge_contract_address"`
				}{
					BridgeWorkers:         []string{"air1h58eezgk5j4jwwpk3nxggx63gfuhnfcj78z5vj"},
					BridgeContractAddress: "0xd47248E2f6C725Dd20C82893162aA545C345834e",
				},
			},
		},
		Metadata:  fmt.Sprintf("ipfs://%s", ipfsCID),
		Deposit:   "51000000uamf",
		Title:     "Update EVM Bridge Authorized Unlockers",
		Summary:   "This proposal aims to update the EVM bridge authorized unlockers list and add new bridge contract addresses to enhance the bridge's security and functionality.",
		Expedited: true,
	}

	proposalData, err := json.MarshalIndent(proposal, "", " ")
	if err != nil {
		fmt.Printf("Error marshaling proposal: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile("proposal.json", proposalData, 0644); err != nil {
		fmt.Printf("Error writing proposal.json: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ proposal.json created successfully")

	// Step 3: Submit proposal to chain
	fmt.Println("\n🚀 Submitting proposal to chain...")
	submitCmd := exec.Command(
		config.JunctiondPath,
		"tx", "gov", "submit-proposal", "proposal.json",
		"--from", config.KeyName,
		"--chain-id", config.ChainID,
		"--fees", "50uamf",
		"--gas", "auto",
		"--keyring-backend", "os",
		"-y",
	)

	if err := runCommand(submitCmd); err != nil {
		fmt.Printf("Error submitting proposal: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Proposal submitted successfully!")
	fmt.Println("\n🎯 Next steps:")
	fmt.Println("1. Wait for the deposit period to end")
	fmt.Println("2. Use 'junction-bridge vote <proposal-id> <vote-option>' to vote")
	fmt.Println("3. Use 'junction-bridge monitor-proposals' to monitor status")
}

func runVote(cmd *cobra.Command, args []string) {
	// Load configuration
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Error reading config file: %v\n", err)
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("Error unmarshaling config: %v\n", err)
		os.Exit(1)
	}

	proposalID := args[0]
	voteOption := args[1]

	// Validate vote option
	validOptions := []string{"yes", "no", "abstain", "no_with_veto"}
	isValid := false
	for _, option := range validOptions {
		if voteOption == option {
			isValid = true
			break
		}
	}

	if !isValid {
		fmt.Printf("Invalid vote option: %s. Valid options are: %s\n", voteOption, strings.Join(validOptions, ", "))
		os.Exit(1)
	}

	fmt.Printf("🗳️  Voting %s on proposal %s...\n", voteOption, proposalID)

	voteCmd := exec.Command(
		config.JunctiondPath,
		"tx", "gov", "vote", proposalID, voteOption,
		"--from", config.KeyName,
		"--chain-id", config.ChainID,
		"--fees", "50uamf",
		"--keyring-backend", "os",
		"-y",
	)

	if err := runCommand(voteCmd); err != nil {
		fmt.Printf("Error voting on proposal: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Successfully voted %s on proposal %s!\n", voteOption, proposalID)
}

func runMonitorProposals(cmd *cobra.Command, args []string) {
	// Load configuration
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Printf("Error reading config file: %v\n", err)
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("Error unmarshaling config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🔍 Monitoring governance proposals...")
	fmt.Println("Press Ctrl+C to stop monitoring")

	// Animation frames for different states
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinnerIndex := 0

	for {
		// Fetch proposals
		proposals, err := fetchProposals(config.RestEndpoint)
		if err != nil {
			fmt.Printf("\r❌ Error fetching proposals: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Clear screen and show status
		fmt.Print("\033[2J\033[H") // Clear screen
		fmt.Println("🔍 Governance Proposals Monitor")
		fmt.Println("================================")

		if len(proposals.Proposals) == 0 {
			fmt.Printf("\r%s No proposals found", spinner[spinnerIndex%len(spinner)])
		} else {
			for _, proposal := range proposals.Proposals {
				status := getStatusDisplay(proposal.Status)
				fmt.Printf("📋 Proposal #%s - %s\n", proposal.ID, status)

				if proposal.Status == "PROPOSAL_STATUS_VOTING_PERIOD" {
					fmt.Printf("   ⏰ Voting Period: %s to %s\n",
						formatTime(proposal.VotingStartTime),
						formatTime(proposal.VotingEndTime))

					// Check if voting period has ended
					if isVotingPeriodEnded(proposal.VotingEndTime) {
						fmt.Println("   🎉 VOTING PERIOD COMPLETED!")
						showCompletionAnimation()
						return
					}
				}

				fmt.Printf("   📊 Tally: Yes: %s, No: %s, Abstain: %s, No with Veto: %s\n",
					proposal.FinalTallyResult.YesCount,
					proposal.FinalTallyResult.NoCount,
					proposal.FinalTallyResult.AbstainCount,
					proposal.FinalTallyResult.NoWithVetoCount)
				fmt.Println()
			}
		}

		spinnerIndex++
		time.Sleep(2 * time.Second)
	}
}

func fetchProposals(restEndpoint string) (*ProposalResponse, error) {
	url := fmt.Sprintf("%s/cosmos/gov/v1/proposals?proposal_status=PROPOSAL_STATUS_UNSPECIFIED", restEndpoint)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var proposalResponse ProposalResponse
	if err := json.Unmarshal(body, &proposalResponse); err != nil {
		return nil, err
	}

	return &proposalResponse, nil
}

func getStatusDisplay(status string) string {
	switch status {
	case "PROPOSAL_STATUS_DEPOSIT_PERIOD":
		return "💰 Deposit Period"
	case "PROPOSAL_STATUS_VOTING_PERIOD":
		return "🗳️  Voting Period"
	case "PROPOSAL_STATUS_PASSED":
		return "✅ PASSED"
	case "PROPOSAL_STATUS_REJECTED":
		return "❌ REJECTED"
	case "PROPOSAL_STATUS_FAILED":
		return "💥 FAILED"
	default:
		return fmt.Sprintf("❓ %s", status)
	}
}

func formatTime(timeStr string) string {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return timeStr
	}
	return t.Format("2006-01-02 15:04:05")
}

func isVotingPeriodEnded(votingEndTime string) bool {
	endTime, err := time.Parse(time.RFC3339, votingEndTime)
	if err != nil {
		return false
	}
	return time.Now().After(endTime)
}

func showCompletionAnimation() {
	fmt.Println("\n🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉")
	fmt.Println("🎉                                               🎉")
	fmt.Println("🎉           PROPOSAL COMPLETED!                🎉")
	fmt.Println("🎉                                               🎉")
	fmt.Println("🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉")

	// Animate the completion message
	for i := 0; i < 5; i++ {
		fmt.Print("\r🎉 PROPOSAL COMPLETED! 🎉")
		time.Sleep(500 * time.Millisecond)
		fmt.Print("\r                    ")
		time.Sleep(500 * time.Millisecond)
	}
	fmt.Println("\r🎉 PROPOSAL COMPLETED! 🎉")
}
