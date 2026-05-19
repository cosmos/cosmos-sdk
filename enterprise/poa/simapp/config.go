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

package simapp

import (
	"slices"

	sigtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
)

type PoAConfig struct {
	AppName                   string
	EnableOptimisticExecution bool
	EnableSecp256k1Support    bool
	EnabledSignModes          []sigtypes.SignMode
	ModuleAccountPermissions  map[string][]string
}

func DefaultPoAConfig() PoAConfig {
	return PoAConfig{
		AppName:                   appName,
		EnableOptimisticExecution: true,
		EnableSecp256k1Support:    true,
		EnabledSignModes:          append(slices.Clone(authtx.DefaultSignModes), sigtypes.SignMode_SIGN_MODE_TEXTUAL),
		ModuleAccountPermissions: map[string][]string{
			authtypes.FeeCollectorName: nil,
			govtypes.ModuleName:        {authtypes.Burner, authtypes.Staking},
			poatypes.ModuleName:        nil,
		},
	}
}
