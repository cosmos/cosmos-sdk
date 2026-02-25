# Group Cosmos SDK Module

> **Part of Cosmos SDK Enterprise Modules** | [Enterprise Modules](../README.md)
>
> **Full Documentation**: [docs/architecture.md](./docs/architecture.md) | [API Reference](./docs/api.md)
>
> **License Notice**: This module uses the [Source Available Evaluation License](./LICENSE), different from the core SDK's Apache-2.0 license. See the [License](#license) section for details.

A Cosmos SDK module that enables on-chain multisig accounts and collective decision-making through configurable voting policies.

## Overview

The Group module allows any set of accounts to form a named group, attach one or more decision policies to it, and collectively authorize the execution of arbitrary messages through a proposal-and-vote workflow.

### Key Features

- **Flexible Membership**: Groups aggregate accounts with weighted voting power; members can be added, removed, or reweighted by the group admin
- **Multiple Decision Policies**: Each group can have multiple group policy accounts, each with its own threshold or percentage-based decision policy
- **Proposal Execution**: When a proposal is accepted according to its policy, any account can trigger execution of the embedded messages
- **Automatic Tally**: At the end of every block, proposals whose voting period has expired are tallied and pruned automatically

## Architecture

The Group module interacts with two core Cosmos SDK modules:

```
                        ┌──────────────────────────────────────────┐
                        │              x/group module              │
                        │                                          │
  Members ──submit──▶   │  ┌─────────┐    ┌──────────────────┐   │
  Members ──vote────▶   │  │  Group  │───▶│   Group Policy   │   │
                        │  │         │    │  (decision policy)│   │
                        │  └─────────┘    └────────┬─────────┘   │
                        │                          │              │
                        │               ┌──────────▼──────────┐  │
                        │               │      Proposal        │  │
                        │               │  (messages + votes)  │  │
                        │               └──────────┬──────────┘  │
                        │                          │              │
                        │               ┌──────────▼──────────┐  │
                        │               │  Message Execution   │  │
                        │               │  (via msg router)    │  │
                        │               └─────────────────────┘  │
                        └──────────────────────────────────────────┘
                                   │                  │
                            ┌──────▼──────┐   ┌───────▼──────┐
                            │   x/auth    │   │   x/bank     │
                            │  (accounts, │   │  (spendable  │
                            │  addr codec)│   │   balances)  │
                            └─────────────┘   └──────────────┘
```

### Module Interactions

| Module | Purpose | Interface |
|--------|---------|-----------|
| **Auth** | Account management | Creates group policy accounts; provides address codec |
| **Bank** | Balance checks | Queried for spendable coins on group policy accounts |
| **BaseApp** | Message routing | Executes proposal messages via the message router |
| **EndBlock** | Lifecycle management | Tallies proposals at voting period end; prunes expired proposals |

## Quick Start

### Prerequisites

- Go 1.25+
- Docker (for proto generation)

### Build

```bash
# Build the module and simd binary
make build

# Or with go directly
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

## Usage

### Query Commands

**Get info about a group by ID:**
```bash
simd q group group-info [group-id]
```

**Get info about a group policy by account address:**
```bash
simd q group group-policy-info [group-policy-account]
```

**List members of a group:**
```bash
simd q group group-members [group-id]
```

**List groups by admin address:**
```bash
simd q group groups-by-admin [admin]
```

**List group policies for a group:**
```bash
simd q group group-policies-by-group [group-id]
```

**Get a proposal by ID:**
```bash
simd q group proposal [proposal-id]
```

**List proposals for a group policy:**
```bash
simd q group proposals-by-group-policy [group-policy-account]
```

**Get a vote:**
```bash
simd q group vote [proposal-id] [voter]
```

**Get the current tally for a proposal:**
```bash
simd q group tally-result [proposal-id]
```

**List all groups on chain:**
```bash
simd q group groups
```

### Transaction Commands

**Create a group:**
```bash
simd tx group create-group [admin] [metadata] [members-json-file]
```

Where `members.json` contains:
```json
{
  "members": [
    {
      "address": "cosmos1...",
      "weight": "1",
      "metadata": "member description"
    }
  ]
}
```

**Create a group policy with a threshold decision policy:**
```bash
simd tx group create-group-policy [admin] [group-id] [metadata] [decision-policy-json]
```

Where the threshold decision policy JSON is:
```json
{
  "@type": "/cosmos.group.v1.ThresholdDecisionPolicy",
  "threshold": "2",
  "windows": {
    "voting_period": "24h",
    "min_execution_period": "0s"
  }
}
```

**Submit a proposal:**
```bash
simd tx group submit-proposal [proposal-json-file] \
    --from proposer \
    --keyring-backend test
```

**Vote on a proposal:**
```bash
simd tx group vote [proposal-id] [voter] [vote-option] [metadata]
# vote-option: VOTE_OPTION_YES | VOTE_OPTION_NO | VOTE_OPTION_ABSTAIN | VOTE_OPTION_NO_WITH_VETO
```

**Execute an accepted proposal:**
```bash
simd tx group exec [proposal-id] \
    --from executor \
    --keyring-backend test
```

**Update group members (set weight to "0" to remove):**
```bash
simd tx group update-group-members [admin] [group-id] [members-json-file]
```

**Leave a group:**
```bash
simd tx group leave-group [member-address] [group-id]
```

**Withdraw a submitted proposal:**
```bash
simd tx group withdraw-proposal [proposal-id] [group-policy-admin-or-proposer]
```

## Testing & Development

### Run Tests

**Unit tests:**
```bash
go test ./x/group/...
```

**Keeper tests:**
```bash
go test ./x/group/keeper/...
```

**All tests with coverage:**
```bash
make test-cover
```

**Fuzz tests:**
```bash
make test-fuzz
```

**System tests:**
```bash
make test-system
```

## Configuration

### Genesis Configuration

The group module starts with an empty genesis state by default. Groups are created at runtime through transactions.

```json
{
  "group": {
    "group_seq": "0",
    "groups": [],
    "group_members": [],
    "group_policy_seq": "0",
    "group_policies": [],
    "proposal_seq": "0",
    "proposals": [],
    "votes": []
  }
}
```

### Module Parameters

The group module is configured at initialization time (not via on-chain governance):

| Parameter | Default | Description |
|-----------|---------|-------------|
| `MaxExecutionPeriod` | `336h` (2 weeks) | Max duration after a proposal's voting period ends that members can execute it |
| `MaxMetadataLen` | `255` | Maximum byte length for metadata fields on groups, policies, proposals, and votes |

## Features In Detail

### Groups and Membership

- **Weighted Members**: Each member has a decimal weight; the group's total weight is the sum of all member weights
- **Admin Control**: A group admin can add, remove, and update members; the admin can also transfer admin rights
- **Dynamic Updates**: Member weights can be changed at any time; active proposals automatically see updated totals at tally time

### Group Policies and Decision Policies

- **Multiple Policies Per Group**: A group can have many group policy accounts, each independently configured
- **Threshold Policy**: A proposal passes when the sum of YES vote weights meets or exceeds a defined threshold
- **Percentage Policy**: A proposal passes when the YES percentage of total group weight meets or exceeds a defined percentage
- **Extensible**: Custom decision policies can be registered by implementing the `DecisionPolicy` interface

### Proposal Lifecycle

1. Any group member submits a proposal containing one or more messages
2. Members vote `YES`, `NO`, `ABSTAIN`, or `NO_WITH_VETO` during the voting window
3. At voting period end (EndBlock), the keeper tallies votes and marks proposals `ACCEPTED` or `REJECTED`
4. Accepted proposals can be executed by any account within `MaxExecutionPeriod`
5. After `MaxExecutionPeriod` elapses, expired proposals are pruned from state

### Message Execution

- Accepted proposals execute their embedded messages through the BaseApp message router
- The group policy account address is used as the signer for executed messages
- Execution can be attempted immediately after voting with `--exec try`, or separately via `exec`

## Development

### Project Structure

```
enterprise/group/
├── proto/cosmos/group/v1/    # Protobuf definitions
├── api/cosmos/group/v1/      # Generated pulsar/grpc files
├── x/group/                  # Module implementation
│   ├── keeper/
│   │   ├── keeper.go         # Keeper initialization and ORM setup
│   │   ├── msg_server.go     # Transaction handlers
│   │   ├── grpc_query.go     # Query handlers
│   │   ├── tally.go          # Vote tallying logic
│   │   └── proposal_executor.go # Message execution
│   ├── client/cli/           # CLI commands
│   ├── module/
│   │   ├── module.go         # AppModule implementation
│   │   ├── abci.go           # EndBlock logic
│   │   └── autocli.go        # AutoCLI query/tx descriptors
│   └── internal/
│       └── orm/              # Object-Relational Mapping layer
├── simapp/                   # Test application with group module
└── tests/systemtests/        # Black-box system tests
```

### Adding New Features

1. Define protobuf messages in `proto/`
2. Generate code: `make proto-gen`
3. Implement message handler in `keeper/msg_server.go`
4. Add query handler in `keeper/grpc_query.go`
5. Register CLI commands in `client/cli/tx.go` or via `module/autocli.go`
6. Write tests in `keeper/*_test.go`

## License

**IMPORTANT**: This module uses a different license than the core Cosmos SDK.

This module is licensed under the **Source Available Evaluation License** for non-commercial evaluation, testing, and educational purposes only. Commercial use requires a separate license.

- **Group Module License**: [Source Available Evaluation License](./LICENSE) - Evaluation/testing only
