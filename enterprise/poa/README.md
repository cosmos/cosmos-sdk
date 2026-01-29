# Proof of Authority (PoA) Cosmos SDK Module

A Cosmos SDK module that implements a Proof of Authority (PoA) consensus mechanism, allowing a designated admin to manage a permissioned validator set and integrate with governance for validator-only participation.

## Overview

The PoA module replaces traditional staking-based validator selection with an admin-controlled validator set. This is useful for permissioned networks, consortium chains, and testing environments where validator participation needs to be tightly controlled.

### Key Features

- **Admin-Controlled Validator Set**: A single admin account manages validator additions, removals, and power adjustments (this admin could be a group, a multisig or the governance module address)
- **Fee Distribution**: Validators accumulate transaction fees proportional to their voting power
- **Governance Integration**: Restricts proposal submission, deposits, and voting to active PoA validators only
- **Custom Vote Tallying**: Governance votes are weighted by validator power instead of staked tokens

## Architecture

The PoA module integrates with several core Cosmos SDK modules and CometBFT:

![diagram](./docs/architecture.png)

### Module Interactions

| Module | Purpose | Interface |
|--------|---------|-----------|
| **CometBFT** | Consensus engine | Receives validator updates via ABCI (BeginBlock/EndBlock) |
| **Auth** | Account management | Gets module accounts and address codec |
| **Bank** | Token transfers | Sends accumulated fees from PoA module to validators |
| **Governance** | Proposals & voting | Hooks validate only PoA validators can participate; custom vote tallying by validator power |

## Breaking Changes

### Cosmos SDK v0.54.x Integration

When integrating with SDK v0.54.x+ (see upgrade guide [here](https://github.com/cosmos/cosmos-sdk/blob/main/UPGRADING.md)), note these breaking changes:

- **x/gov Keeper Initialization** ([#25615](https://github.com/cosmos/cosmos-sdk/pull/25615)): `keeper.NewKeeper` now requires a `CalculateVoteResultsAndVotingPowerFn` parameter instead of `StakingKeeper`. Use `keeper.NewDefaultCalculateVoteResultsAndVotingPower(stakingKeeper)` to wrap your staking keeper.

- **x/gov GovHooks Interface** ([#25617](https://github.com/cosmos/cosmos-sdk/pull/25617)): `AfterProposalSubmission` hook now includes `proposerAddr sdk.AccAddress` as an additional parameter.

- **x/gov Distribution Keeper** ([#25616](https://github.com/cosmos/cosmos-sdk/pull/25616)): `DistrKeeper` is now optional, but must be non-nil if used as a cancellation fee destination.

See the full [Cosmos SDK v0.54.x Changelog](https://github.com/cosmos/cosmos-sdk/blob/main/CHANGELOG.md) for details.

## Quick Start

### Prerequisites

- Go 1.25+
- Docker (for localnet)

### Local Development Network

Start a local test network with the PoA module:

```bash
# Start localnet (rebuilds chain, resets genesis, starts nodes)
make localnet

# Access node1 shell for testing
make localnet-shell

# Clean up localnet
make localnet-clean
```

## Usage

All commands below should be run inside `make localnet-shell` or with appropriate node configuration.

### Query Commands

**Get PoA parameters (including admin address):**
```bash
simd q poa params
```

**List all validators and their power:**
```bash
simd q poa validators
```

**Get specific validator by consensus address:**
```bash
simd q poa validator <consensus-address>
```

### Transaction Commands

**Update PoA parameters (admin only):**
```bash
simd tx poa update-params \
    --admin cosmos1x0mm8rws8lm46xay3zyyznzr6lvu5um3kht0x7 \
    --from admin \
    --keyring-backend test
```

**Update validators (admin only):**

Add or modify validators by specifying their public key and power:
```bash
simd tx poa update-validators validators.json \
    --from admin \
    --keyring-backend test
```

**Withdraw validator fees:**
```bash
simd tx poa withdraw-fees \
    --from account \
    --keyring-backend test
```

### Creating and Adding a New Validator

1. **Generate validator operator key and fund the account:**
```bash
# Add a new key for your validator operator (this is for signing transactions, not consensus)
simd keys add myvalidator --keyring-backend test

export OPERATOR_ADDR="$(simd keys show myvalidator --keyring-backend=test --address)"

# Fund the account from an existing account
simd tx bank send account $OPERATOR_ADDR 100000token \
  --from account \
  --keyring-backend test \
  --chain-id poa-localnet-1 \
  --fees=100token \
  -y
```

2. **Get the validator consensus public key (ed25519):**

**Important:** The account key from `simd keys add` (secp256k1) is for signing transactions, not for validator consensus. You need a separate ed25519 consensus public key.

```bash
# Option A: If you have a running validator node, get its consensus pubkey:
simd tendermint show-validator

export CONS_PUBKEY="$(simd comet show-validator | jq -r '.key')"

# Option B: For testing, you can use an arbitrary ed25519 key like 13iyxnnVneLg0AxHeUD7dRAegA8W3gB1mT4p7sPGjyY=
```

3. **Create the validator (any funded account can do this):**
```bash
# Create validator with zero power (pending activation by admin)
simd tx poa create-validator \
    "My Validator" \
    $CONS_PUBKEY \
    "ed25519" \
    --description "My validator description" \
    --from myvalidator \
    --keyring-backend test \
    --fees 1token
```

4. **Admin activates the validator by setting power:**

Create a `validators.json` file:
```json
[
  {
    "pub_key": {
      "@type": "/cosmos.crypto.ed25519.PubKey",
      "key": "<base64-encoded-ed25519-pubkey>"
    },
    "power": 10000,
    "metadata": {
      "moniker": "My Validator",
      "description": "My validator description",
      "operator_address": "cosmos1..."
    }
  }
]
```

Then update validators:
```bash
simd tx poa update-validators validators.json \
    --from account \
    --keyring-backend test
```

**Notes:**
- Setting `power` to `0` removes a validator from the active set
- The `operator_address` is the address that can withdraw fees

## Testing & Development

### Network Testing

**Check node status:**
```bash
simd status
```

**Get current block height:**
```bash
simd q block | jq '.block.header.height'
```

**View recent blocks:**
```bash
simd q block <height>
```

### Governance Testing

**Test governance (validators only):**
```bash
# Submit a proposal (must be from a validator account)
simd tx gov submit-proposal proposal.json \
    --from account \
    --keyring-backend test

# Vote on proposal (only authorized validators can vote)
simd tx gov vote 1 yes \
    --from account \
    --keyring-backend test

# Check proposal status
simd q gov proposal 1
```

### Integration Testing

**Run unit tests:**
```bash
go test ./x/poa/...
```

**Run keeper tests:**
```bash
go test ./x/poa/keeper/...
```

**Run all tests with coverage:**
```bash
go test -coverprofile=coverage.out ./x/poa/...
go tool cover -html=coverage.out
```

## Configuration

### Genesis Configuration

The PoA module requires the following genesis configuration:

```json
{
  "poa": {
    "params": {
      "admin": "cosmos1..."
    },
    "validators": [
      {
        "pub_key": {
          "@type": "/cosmos.crypto.ed25519.PubKey",
          "key": "base64-encoded-pubkey"
        },
        "power": 10000000,
        "metadata": {
          "operator_address": "cosmos1...",
          "moniker": "validator1"
        }
      }
    ]
  }
}
```

### Module Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `admin` | string | Bech32 address of the admin account that can update validators and params |

## Features In Detail

### Validator Management

- **Admin-Only Control**: Only the designated admin account can add, remove, or modify validators
- **Power-Based Ordering**: Validators are indexed by power for efficient iteration
- **Consensus & Operator Addresses**: Each validator has both a consensus address (for block signing) and an operator address (for transactions)

### Fee Distribution

- Validators automatically accumulate transaction fees proportional to their voting power
- Fees are stored as `DecCoins` to handle fractional amounts precisely
- Validators withdraw fees to their operator address via `withdraw-fees` transaction

### Governance Integration

- **Restricted Participation**: Only active PoA validators with power > 0 can submit proposals, deposit, and vote
- **Power-Based Voting**: Votes are tallied using validator power instead of token stakes
- **Governance Hooks**: Validates voter authorization at proposal submission, deposit, and vote time

## Development

### Project Structure

```
x/poa/
├── keeper/          # Core business logic
│   ├── keeper.go       # Keeper initialization
│   ├── validator.go    # Validator management
│   ├── distribution.go # Fee distribution
│   ├── governance.go   # Governance integration
│   ├── hooks.go        # Module hooks
│   └── abci.go         # BeginBlock/EndBlock
├── types/           # Type definitions
│   ├── poa.proto       # Protobuf definitions
│   ├── expected_keepers.go # Keeper interfaces
│   └── keys.go         # Storage keys
└── client/cli/      # CLI commands

```

### Adding New Features

1. Define protobuf messages in `proto/`
2. Generate code: `make proto-gen`
3. Implement keeper methods in `keeper/`
4. Add CLI commands in `client/cli/`
5. Write tests in `keeper/*_test.go`

## License

See [LICENSE](LICENSE) file for details.
