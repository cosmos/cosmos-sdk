// IMPORTANT LICENSE NOTICE
//
// SPDX-License-Identifier: CosmosLabs-Evaluation-Only
//
// This file is NOT licensed under the Apache License 2.0.
//
// Licensed under the Cosmos Labs Source Available Evaluation License, which forbids:
// - commercial use,
// - production use, and
// - redistribution.
//
// See https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/poa/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

package types

import (
	"cosmossdk.io/errors"
)

var (
	ErrInvalidSigner                  = errors.Register(ModuleName, 0, "invalid signer for message, expected admin")
	ErrNegativeValidatorPower         = errors.Register(ModuleName, 1, "negative validator power, validator powers must be non-negative")
	ErrMissingOperatorAddress         = errors.Register(ModuleName, 2, "missing validator operator address, all validators must have an associated validator operator address")
	ErrUnknownValidator               = errors.Register(ModuleName, 3, "attempted to update unknown validator, all validators must be created before being updated")
	ErrInvalidMetadata                = errors.Register(ModuleName, 4, "invalid metadata")
	ErrInvalidTotalPower              = errors.Register(ModuleName, 5, "invalid total power")
	ErrValidatorAlreadyExists         = errors.Register(ModuleName, 6, "validator already exists, cannot create duplicate validator with same consensus address")
	ErrSameKeyForOperatorAndConsensus = errors.Register(ModuleName, 7, "operator address and consensus pubkey must use different keys to prevent accidental key reuse")
	ErrInvalidAdminAddress            = errors.Register(ModuleName, 8, "invalid admin address")
)
