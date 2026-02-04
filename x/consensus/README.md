---
sidebar_position: 1
---

# `x/consensus`

The consensus module provides functionality to modify CometBFT's ABCI consensus parameters on-chain through governance proposals.

## Overview

This module allows authorized entities (typically governance) to update critical consensus parameters that affect blockchain performance and security without requiring a hard fork.

## Key Features

- **Parameter Updates**: Modify consensus parameters through governance proposals
- **Authority Control**: Only authorized addresses can update parameters
- **Validation**: Comprehensive parameter validation before updates
- **ABCI Integration**: Direct integration with CometBFT consensus engine

## Consensus Parameters

The module manages the following CometBFT consensus parameters:

### Block Parameters
- `MaxBytes`: Maximum block size in bytes
- `MaxGas`: Maximum gas per block

### Evidence Parameters
- `MaxAgeNumBlocks`: Maximum age of evidence in blocks
- `MaxAgeDuration`: Maximum age of evidence in time
- `MaxBytes`: Maximum evidence size in bytes

### Validator Parameters
- `PubKeyTypes`: Supported public key types for validators

### ABCI Parameters
- `VoteExtensionsEnableHeight`: Height at which vote extensions are enabled

## Usage

### Governance Proposal

To update consensus parameters, submit a governance proposal with `MsgUpdateParams`:

```go
import (
    "time"
    govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
    authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

msg := &types.MsgUpdateParams{
    Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(), // governance authority
    Block: &types.BlockParams{
        MaxBytes: 22020096,
        MaxGas:   10000000,
    },
    Evidence: &types.EvidenceParams{
        MaxAgeNumBlocks: 100000,
        MaxAgeDuration:  48 * time.Hour,
        MaxBytes:        1048576,
    },
    Validator: &types.ValidatorParams{
        PubKeyTypes: []string{"ed25519"},
    },
    Abci: &types.ABCIParams{
        VoteExtensionsEnableHeight: 0,
    },
}
```

### Query Parameters

Retrieve current consensus parameters:

```bash
<appd> q consensus params
```

## Architecture

- **Keeper**: Manages parameter storage and validation
- **Types**: Defines message types and parameter structures
- **Module**: Integrates with the Cosmos SDK application lifecycle

## Security

- Only the designated authority can update parameters
- All parameter changes are validated before application
- Updates are subject to governance approval process
