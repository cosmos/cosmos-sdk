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
	"slices"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log/v2"
	"cosmossdk.io/math"

	sdkapp "github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa"
	poakeeper "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/keeper"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const appName = "SimApp"

var DefaultNodeHome string

var _ sdkapp.AppI = (*SimApp)(nil)

type SimApp struct {
	*sdkapp.SDKApp

	POAKeeper *poakeeper.Keeper
}

func init() {
	var err error
	DefaultNodeHome, err = sdkapp.GetNodeHomeDirectory(".simapp")
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
	poaConfig := DefaultPoAConfig()
	var poaKeeper *poakeeper.Keeper

	sdkAppConfig := sdkapp.DefaultSDKAppConfig(poaConfig.AppName, appOpts, baseAppOptions...)
	sdkAppConfig.WithAuthz = false
	sdkAppConfig.WithEpochs = false
	sdkAppConfig.WithFeeGrant = false
	sdkAppConfig.OptimisticExecutionEnabled = poaConfig.EnableOptimisticExecution
	sdkAppConfig.Keys = append(slices.Clone(sdkAppConfig.Keys), poatypes.StoreKey)
	sdkAppConfig.TransientStoreKeys = append(slices.Clone(sdkAppConfig.TransientStoreKeys), poatypes.TransientStoreKey)
	sdkAppConfig.ModuleAccountPerms[govtypes.ModuleName] = []string{authtypes.Burner, authtypes.Staking}
	sdkAppConfig.ModuleAccountPerms[poatypes.ModuleName] = nil
	sdkAppConfig.OrderBeginBlockers = append(slices.Clone(sdkAppConfig.OrderBeginBlockers), poatypes.ModuleName)
	sdkAppConfig.OrderEndBlockers = []string{
		genutiltypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		poatypes.ModuleName,
		banktypes.ModuleName,
	}
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
	sdkAppConfig.GovHooks = []govtypes.GovHooks{
		deferredPOAGovHooks{poaKeeper: &poaKeeper},
	}

	sharedApp := sdkapp.NewSDKApp(logger, db, nil, sdkAppConfig)
	poaKeeper = poakeeper.NewKeeper(
		sharedApp.AppCodec(),
		runtime.NewKVStoreService(sharedApp.GetKey(poatypes.StoreKey)),
		runtime.NewTransientStoreService(sharedApp.GetTransientStoreKey(poatypes.TransientStoreKey)),
		sharedApp.AccountKeeper,
		sharedApp.BankKeeper,
	)

	poaAppModule := poa.NewAppModule(sharedApp.AppCodec(), poaKeeper)
	if poaConfig.EnableSecp256k1Support {
		poaAppModule = poa.NewAppModule(sharedApp.AppCodec(), poaKeeper, poa.WithSecp256k1Support())
	}

	if err := sharedApp.AddModules(poaModule{AppModule: poaAppModule}); err != nil {
		panic(err)
	}
	sharedApp.LoadModules()
	anteHandler, err := NewAnteHandler(ante.HandlerOptions{
		AccountKeeper:   sharedApp.AccountKeeper,
		BankKeeper:      sharedApp.BankKeeper,
		FeegrantKeeper:  sharedApp.FeeGrantKeeper,
		SignModeHandler: sharedApp.TxConfig().SignModeHandler(),
		SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
	})
	if err != nil {
		panic(err)
	}
	sharedApp.SetAnteHandler(anteHandler)

	simApp := &SimApp{
		SDKApp:    sharedApp,
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

type deferredPOAGovHooks struct {
	poaKeeper **poakeeper.Keeper
}

func (h deferredPOAGovHooks) getHooks() (govtypes.GovHooks, error) {
	if h.poaKeeper == nil || *h.poaKeeper == nil {
		return nil, errors.New("poa keeper is not initialized")
	}

	return (*h.poaKeeper).NewGovHooks(), nil
}

func (h deferredPOAGovHooks) AfterProposalSubmission(ctx context.Context, proposalID uint64, proposerAddr sdk.AccAddress) error {
	hooks, err := h.getHooks()
	if err != nil {
		return err
	}
	return hooks.AfterProposalSubmission(ctx, proposalID, proposerAddr)
}

func (h deferredPOAGovHooks) AfterProposalDeposit(ctx context.Context, proposalID uint64, depositorAddr sdk.AccAddress) error {
	hooks, err := h.getHooks()
	if err != nil {
		return err
	}
	return hooks.AfterProposalDeposit(ctx, proposalID, depositorAddr)
}

func (h deferredPOAGovHooks) AfterProposalVote(ctx context.Context, proposalID uint64, voterAddr sdk.AccAddress) error {
	hooks, err := h.getHooks()
	if err != nil {
		return err
	}
	return hooks.AfterProposalVote(ctx, proposalID, voterAddr)
}

func (h deferredPOAGovHooks) AfterProposalFailedMinDeposit(ctx context.Context, proposalID uint64) error {
	hooks, err := h.getHooks()
	if err != nil {
		return err
	}
	return hooks.AfterProposalFailedMinDeposit(ctx, proposalID)
}

func (h deferredPOAGovHooks) AfterProposalVotingPeriodEnded(ctx context.Context, proposalID uint64) error {
	hooks, err := h.getHooks()
	if err != nil {
		return err
	}
	return hooks.AfterProposalVotingPeriodEnded(ctx, proposalID)
}
