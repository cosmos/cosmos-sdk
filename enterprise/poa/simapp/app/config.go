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

package app

import (
	"slices"

	sdkapp "github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/baseapp"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sigtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const DefaultAppName = "SimApp"

type PoAConfig struct {
	AppName                   string
	EnableOptimisticExecution bool
	EnableSecp256k1Support    bool
	EnabledSignModes          []sigtypes.SignMode
	ModuleAccountPermissions  map[string][]string
}

func DefaultPoAConfig() PoAConfig {
	return PoAConfig{
		AppName:                   DefaultAppName,
		EnableOptimisticExecution: true,
		EnableSecp256k1Support:    true,
		EnabledSignModes:          append(slices.Clone(authtx.DefaultSignModes), sigtypes.SignMode_SIGN_MODE_TEXTUAL),
		ModuleAccountPermissions: map[string][]string{
			authtypes.FeeCollectorName: nil,
			govtypes.ModuleName:        {authtypes.Burner},
			poatypes.ModuleName:        nil,
		},
	}
}

// NewPoAConfig returns the canonical PoA simapp configuration.
func NewPoAConfig() PoAConfig {
	return DefaultPoAConfig()
}

// DefaultPoAAppConfig returns the canonical SDK app config for PoA.
func DefaultPoAAppConfig(
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) sdkapp.SDKAppConfig {
	return NewPoAAppConfig(NewPoAConfig(), appOpts, baseAppOptions...)
}

// NewPoAAppConfig returns an SDK app config composed from PoA settings.
func NewPoAAppConfig(
	poaConfig PoAConfig,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) sdkapp.SDKAppConfig {
	sdkAppConfig := sdkapp.DefaultSDKAppConfig(poaConfig.AppName, appOpts, baseAppOptions...)
	sdkAppConfig.WithAuthz = false
	sdkAppConfig.WithEpochs = false
	sdkAppConfig.WithFeeGrant = false
	sdkAppConfig.OptimisticExecutionEnabled = poaConfig.EnableOptimisticExecution
	sdkAppConfig.Keys = append(slices.Clone(sdkAppConfig.Keys), poatypes.StoreKey)
	sdkAppConfig.TransientStoreKeys = append(slices.Clone(sdkAppConfig.TransientStoreKeys), poatypes.TransientStoreKey)
	for moduleName, perms := range poaConfig.ModuleAccountPermissions {
		sdkAppConfig.ModuleAccountPerms[moduleName] = slices.Clone(perms)
	}
	sdkAppConfig.OrderBeginBlockers = append(slices.Clone(sdkAppConfig.OrderBeginBlockers), poatypes.ModuleName)
	sdkAppConfig.OrderEndBlockers = []string{
		genutiltypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		poatypes.ModuleName,
		banktypes.ModuleName,
	}

	return sdkAppConfig
}
