# Multi-Chain Testing Framework

A comprehensive collection of automated testing runners for various blockchain networks. Each chain has its own dedicated testing suite with interactive CLI tools, governance testing, and parameter management.

## 🚀 Overview

This repository contains specialized testing runners for different blockchain networks, designed to streamline chain testing, governance proposals, and parameter management across multiple ecosystems.

## 📁 Repository Structure

```
chain_testing_scripts/
├── junction-bridgev1.2.0/     # Junction chain testing suite
├── [chain-name]/              # Individual chain testing directories
│   ├── main.go                # Go-based testing executable
│   ├── build_executable.sh    # Build script
│   ├── env.example           # Environment configuration
│   └── README.md             # Chain-specific documentation
└── README.md                 # This file
```

## 🎯 Features

### **Universal Testing Capabilities**

- **Interactive CLI** with beautiful animations and progress indicators
- **Automated chain initialization** with configurable parameters
- **Governance testing** including proposal creation and voting
- **Parameter management** for bridge workers, contract addresses, and more
- **Environment variable support** for automated testing workflows

### **Chain-Specific Implementations**

Each chain directory contains:

- **Custom executable** tailored to the chain's specific requirements
- **Parameter change proposals** for chain-specific governance
- **Voting mechanisms** with chain-appropriate configurations
- **CLI animations** and user experience optimizations

## 🛠️ Supported Chains

| Chain        | Directory                | Status     | Features                                        |
| ------------ | ------------------------ | ---------- | ----------------------------------------------- |
| Junction     | `junction-bridgev1.2.0/` | ✅ Active  | Bridge testing, parameter proposals, governance |
| [Chain Name] | `[chain-name]/`          | 🔄 Planned | [Features]                                      |

## 🚀 Quick Start

### For Junction Chain

```bash
cd junction-bridgev1.2.0/
./build_executable.sh
./chain-tester
```

### For Other Chains

```bash
cd [chain-name]/
./build_executable.sh
./chain-tester
```

## 🔧 Adding New Chains

To add a new chain testing suite:

1. **Create directory**: `mkdir [chain-name]`
2. **Copy template**: Use existing chain as template
3. **Customize**: Modify parameters, governance, and chain-specific features
4. **Document**: Update chain-specific README
5. **Test**: Ensure all functionality works with the new chain

## 📋 Common Features Across All Chains

- **Chain Initialization**: Automated setup with custom parameters
- **Key Management**: Secure key generation and management
- **Genesis Configuration**: Custom genesis file modifications
- **Governance Testing**: Proposal creation, submission, and voting
- **Parameter Management**: Bridge workers, contract addresses, and more
- **CLI Animations**: Loading spinners, countdown timers, progress indicators
- **Environment Support**: `.env` file configuration for automated testing

## 🎨 User Experience

Each chain runner provides:

- **Step-by-step execution** with clear command previews
- **Interactive prompts** for user input when needed
- **Environment variable fallbacks** for automated workflows
- **Real-time animations** for long-running operations
- **Comprehensive error handling** with helpful messages

## 🔒 Security & Best Practices

- **Environment variable support** for sensitive configuration
- **Secure key management** with proper keyring backends
- **Input validation** for all user-provided data
- **Error handling** with graceful fallbacks
- **Documentation** for each chain's specific requirements

## 📚 Documentation

Each chain directory contains:

- **Chain-specific README** with detailed usage instructions
- **Environment configuration** examples
- **Troubleshooting guides** for common issues
- **Feature documentation** for chain-specific capabilities

## 🤝 Contributing

When adding new chains or improving existing ones:

1. Follow the established directory structure
2. Include comprehensive documentation
3. Test all functionality thoroughly
4. Update this main README with new chain information
5. Ensure consistent user experience across all chains

## 📄 License

This project is part of the multi-chain testing ecosystem for blockchain development and testing.

---

**Ready to test your chains?** Navigate to your desired chain directory and start testing! 🚀
