# Group Cosmos SDK Module

> **Part of Cosmos SDK Enterprise Modules**
>
> **License Notice**: This module uses the [Source Available Evaluation License](./LICENSE), different from the core SDK's Apache-2.0 license.

A Cosmos SDK module that allows the creation and management of on-chain multisig accounts and enables voting for message execution based on configurable decision policies.

## Overview

The Group module enables:
- **Groups**: Aggregations of accounts with associated weights
- **Group Policies**: Accounts associated with a group and a decision policy
- **Proposals**: Members can submit proposals for group policy accounts to decide upon
- **Decision Policies**: Threshold and percentage-based voting (extensible)

## Quick Start

### Prerequisites

- Go 1.25+
- Docker (for proto generation)

### Build

```bash
# Generate protobuf code
make proto-gen

# Build
go build ./...
```

### Test

```bash
# Run unit tests
make test

# Run linter
make lint

# Run fuzz tests (types-level, coverage-guided)
make test-fuzz
```

## Module Structure

```
enterprise/group/
├── proto/cosmos/group/v1/    # Protobuf definitions
├── api/cosmos/group/v1/      # Generated pulsar/grpc files
├── x/group/                  # Module implementation
│   ├── keeper/
│   ├── client/cli/
│   ├── module/
│   ├── internal/
│   └── migrations/
├── simapp/                   # Test application with group module
└── tests/systemtests/        # Black-box system tests
```

## Documentation

- [API Reference](./docs/api.md)
- [Architecture](./docs/architecture.md)

## License

This module is licensed under the **Source Available Evaluation License** for non-commercial evaluation, testing, and educational purposes only.
