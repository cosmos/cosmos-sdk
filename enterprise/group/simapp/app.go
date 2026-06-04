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
// See https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/group/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

package simapp

import (
	"fmt"
	"time"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/enterprise/group/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/enterprise/group/x/group/keeper"
	groupmodule "github.com/cosmos/cosmos-sdk/enterprise/group/x/group/module"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

const appName = "SimApp"

// DefaultNodeHome default home directories for the application daemon.
var DefaultNodeHome string

var _ app.AppI = (*SimApp)(nil)

type SimApp struct {
	*app.SDKApp
	GroupKeeper groupkeeper.Keeper
}

type groupAppModule struct {
	groupmodule.AppModule
	storeKey *storetypes.KVStoreKey
}

func (m groupAppModule) StoreKeys() map[string]*storetypes.KVStoreKey {
	return map[string]*storetypes.KVStoreKey{
		group.StoreKey: m.storeKey,
	}
}

func (groupAppModule) ModuleAccountPermissions() map[string][]string {
	return map[string][]string{}
}

func init() {
	var err error
	DefaultNodeHome, err = app.GetNodeHomeDirectory(".simapp")
	if err != nil {
		panic(err)
	}
}

func NewSimApp(
	logger log.Logger,
	db dbm.DB,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *SimApp {
	sdkAppConfig := app.DefaultSDKAppConfig(appName, appOpts, baseAppOptions...)

	// Keep the Group simapp close to previous behavior by disabling unused optional modules.
	sdkAppConfig.WithAuthz = false
	sdkAppConfig.WithEpochs = false
	sdkAppConfig.WithFeeGrant = false
	sdkAppConfig.ModuleAuthority = authtypes.NewModuleAddress(group.ModuleName).String()

	// Keep custom group ordering local to this constructor.
	sdkAppConfig.OrderBeginBlockers = []string{
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		genutiltypes.ModuleName,
	}
	sdkAppConfig.OrderEndBlockers = []string{
		genutiltypes.ModuleName,
		group.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		banktypes.ModuleName,
	}
	sdkAppConfig.OrderInitGenesis = []string{
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		consensusparamtypes.ModuleName,
		group.ModuleName,
	}
	sdkAppConfig.OrderExportGenesis = []string{
		consensusparamtypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		group.ModuleName,
	}
	sdkApp := app.NewSDKApp(logger, db, nil, sdkAppConfig)

	groupConfig := group.DefaultConfig()
	groupConfig.MaxExecutionPeriod = 14 * 24 * time.Hour
	groupConfig.MaxMetadataLen = 255

	groupStoreKey := storetypes.NewKVStoreKey(group.StoreKey)
	groupKeeper := groupkeeper.NewKeeper(
		groupStoreKey,
		sdkApp.EncodingConfig().Codec,
		sdkApp.MsgServiceRouter(),
		sdkApp.AccountKeeper,
		groupConfig,
	)

	groupModule := groupAppModule{
		AppModule: groupmodule.NewAppModule(
			sdkApp.EncodingConfig().Codec,
			groupKeeper,
			sdkApp.AccountKeeper,
			sdkApp.BankKeeper,
			sdkApp.InterfaceRegistry(),
		),
		storeKey: groupStoreKey,
	}
	if err := sdkApp.AddModules(groupModule); err != nil {
		panic(err)
	}

	sdkApp.LoadModules()
	anteHandler, err := NewAnteHandler(ante.HandlerOptions{
		AccountKeeper:   sdkApp.AccountKeeper,
		BankKeeper:      sdkApp.BankKeeper,
		SignModeHandler: sdkApp.TxConfig().SignModeHandler(),
		SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
	})
	if err != nil {
		panic(err)
	}
	sdkApp.SetAnteHandler(anteHandler)

	simApp := &SimApp{
		SDKApp:      sdkApp,
		GroupKeeper: groupKeeper,
	}

	if loadLatest {
		if err := simApp.LoadLatestVersion(); err != nil {
			panic(fmt.Errorf("error loading last version: %w", err))
		}
	}

	return simApp
}
