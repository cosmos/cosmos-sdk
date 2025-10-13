# Cosmos SDK Protocol Buffers

This directory contains the public protocol buffers API definitions for the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk). These protobuf files define the core data structures, services, and APIs used across all Cosmos SDK modules.

## Overview

The Cosmos SDK uses Protocol Buffers extensively for:
- **Message definitions** - Transaction messages and query requests/responses
- **Service definitions** - gRPC services for queries and transactions  
- **Type definitions** - Core data structures like accounts, coins, and governance
- **API contracts** - Standardized interfaces between modules and clients

## Structure

```
proto/
├── cosmos/           # Core Cosmos SDK modules (auth, bank, gov, etc.)
├── tendermint/       # CometBFT/Tendermint types
├── amino/           # Amino encoding definitions
└── buf.*.yaml       # Buf configuration files
```

## Key Components

### Core Modules (`cosmos/`)
- **`auth/`** - Account authentication and transaction signing
- **`bank/`** - Token transfers and balance management
- **`gov/`** - On-chain governance and voting
- **`staking/`** - Proof-of-stake validator management
- **`distribution/`** - Fee and reward distribution
- **`slashing/`** - Validator penalty mechanisms
- **`mint/`** - Token minting and inflation
- **`upgrade/`** - Software upgrade coordination

### Generation Targets
- **Go code** - Generated using gogoproto and pulsar generators
- **Swagger docs** - REST API documentation
- **gRPC services** - Service definitions for queries and transactions

## Usage

These protobuf files are published to [buf.build/cosmos/cosmos-sdk](https://buf.build/cosmos/cosmos-sdk) and can be imported by other projects:

```yaml
deps:
  - buf.build/cosmos/cosmos-sdk
```

For detailed usage instructions, see the [protobuf documentation](https://docs.cosmos.network/main/build/tooling/protobuf).
