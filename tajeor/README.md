# Tajeor (TJR) Blockchain

Tajeor is an EVM-compatible Layer 1 blockchain built on the Cosmos SDK with Ethermint integration. It provides a high-performance, scalable platform for decentralized applications with full Ethereum compatibility.

## Tokenomics

- **Token Name**: TJR
- **Total Supply**: 1,000,000,000 TJR
- **Initial Mint**: 500,000,000 TJR
- **Staking Rewards**: 500,000,000 TJR (reserved for staking rewards)

## Features

- EVM Compatibility
- High Performance
- Scalable Architecture
- Staking and Delegation
- Interoperability with Cosmos Ecosystem
- Full Ethereum Tooling Support

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Git
- Make

### Installation

1. Clone the repository:
```bash
git clone https://github.com/tajeor/chain
cd chain
```

2. Install dependencies:
```bash
make install
```

3. Initialize the chain:
```bash
tajeord init <moniker> --chain-id tajeor-1
```

4. Start the node:
```bash
tajeord start
```

## Development

### Building

```bash
make build
```

### Testing

```bash
make test
```

## Contributing

We welcome contributions from the community. Please read our [Contributing Guidelines](CONTRIBUTING.md) for more information.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 