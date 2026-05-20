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
	"context"
	"errors"
	"fmt"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log/v2"
	"cosmossdk.io/math"

	sdkapp "github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/baseapp"
	poaapp "github.com/cosmos/cosmos-sdk/enterprise/poa/simapp/app"
	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa"
	poakeeper "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/keeper"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

const appName = "SimApp"

var DefaultNodeHome string

var _ sdkapp.AppI = (*SimApp)(nil)

type SimApp struct {
	*sdkapp.SDKApp

	POAKeeper *poakeeper.Keeper
}

type PoAConfig = poaapp.PoAConfig

func init() {
	var err error
	DefaultNodeHome, err = sdkapp.GetNodeHomeDirectory(".simapp")
	if err != nil {
		panic(err)
	}
}

// NewPoAApp builds the PoA simapp using the canonical PoA config.
func NewPoAApp(
	logger log.Logger,
	db dbm.DB,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *SimApp {
	poaConfig := poaapp.NewPoAConfig()
	sdkAppConfig := poaapp.DefaultPoAAppConfig(appOpts, baseAppOptions...)
	return newPoAAppWithSDKConfig(logger, db, loadLatest, poaConfig, sdkAppConfig)
}

// NewPoAAppWithConfig builds a PoA simapp with explicit PoA configuration.
func NewPoAAppWithConfig(
	logger log.Logger,
	db dbm.DB,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	poaConfig PoAConfig,
	baseAppOptions ...func(*baseapp.BaseApp),
) *SimApp {
	sdkAppConfig := poaapp.NewPoAAppConfig(poaConfig, appOpts, baseAppOptions...)
	return newPoAAppWithSDKConfig(logger, db, loadLatest, poaConfig, sdkAppConfig)
}

func newPoAAppWithSDKConfig(
	logger log.Logger,
	db dbm.DB,
	loadLatest bool,
	poaConfig PoAConfig,
	sdkAppConfig sdkapp.SDKAppConfig,
) *SimApp {
	var poaKeeper *poakeeper.Keeper

	// PoA custom tally needs POA keeper, which is initialized right after SDKApp construction.
	sdkAppConfig.GovVoteCalcFn = func(
		ctx context.Context,
		k govkeeper.Keeper,
		proposal govv1.Proposal,
	) (math.LegacyDec, math.Int, map[govv1.VoteOption]math.LegacyDec, error) {
		if poaKeeper == nil {
			return math.LegacyZeroDec(), math.ZeroInt(), nil, errors.New("poa keeper is not initialized")
		}

		return poakeeper.NewPOACalculateVoteResultsAndVotingPowerFn(*poaKeeper)(ctx, k, proposal)
	}

	poaApp := sdkapp.NewSDKApp(logger, db, nil, sdkAppConfig)
	poaKeeper = poakeeper.NewKeeper(
		poaApp.AppCodec(),
		runtime.NewKVStoreService(poaApp.GetKey(poatypes.StoreKey)),
		runtime.NewTransientStoreService(poaApp.GetTransientStoreKey(poatypes.TransientStoreKey)),
		poaApp.AccountKeeper,
		poaApp.BankKeeper,
	)

	poaAppModule := poa.NewAppModule(poaApp.AppCodec(), poaKeeper)
	if poaConfig.EnableSecp256k1Support {
		poaAppModule = poa.NewAppModule(poaApp.AppCodec(), poaKeeper, poa.WithSecp256k1Support())
	}

	if err := poaApp.AddModules(poaModule{AppModule: poaAppModule}); err != nil {
		panic(err)
	}
	poaApp.LoadModules()
	poaApp.GovKeeper.SetHooks(govtypes.NewMultiGovHooks(poaKeeper.NewGovHooks()))
	anteHandler, err := NewAnteHandler(ante.HandlerOptions{
		AccountKeeper:   poaApp.AccountKeeper,
		BankKeeper:      poaApp.BankKeeper,
		FeegrantKeeper:  poaApp.FeeGrantKeeper,
		SignModeHandler: poaApp.TxConfig().SignModeHandler(),
		SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
	})
	if err != nil {
		panic(err)
	}
	poaApp.SetAnteHandler(anteHandler)

	simApp := &SimApp{
		SDKApp:    poaApp,
		POAKeeper: poaKeeper,
	}

	if loadLatest {
		if err := simApp.LoadLatestVersion(); err != nil {
			panic(fmt.Errorf("error loading last version: %w", err))
		}
	}

	return simApp
}

type poaModule struct {
	poa.AppModule
}

func (poaModule) StoreKeys() map[string]*storetypes.KVStoreKey {
	return map[string]*storetypes.KVStoreKey{}
}

func (poaModule) ModuleAccountPermissions() map[string][]string {
	return map[string][]string{}
}
