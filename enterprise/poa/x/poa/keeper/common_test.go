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

package keeper

import (
	"fmt"
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/log/v2"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

type testFixture struct {
	ctx        sdk.Context
	poaKeeper  *Keeper
	govKeeper  *govkeeper.Keeper
	authKeeper authkeeper.AccountKeeper
	bankKeeper bankkeeper.Keeper
	cdc        codec.Codec
}

// BlockedAddresses returns all the app's blocked account addresses.
func BlockedAddresses(maccPerms map[string][]string) map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	// allow the following addresses to receive funds
	delete(modAccAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	return modAccAddrs
}

func setupTest(t *testing.T) *testFixture {
	t.Helper()

	// Setup keys
	storeKey := storetypes.NewKVStoreKey(poatypes.StoreKey)
	govStoreKey := storetypes.NewKVStoreKey(govtypes.StoreKey)
	authStoreKey := storetypes.NewKVStoreKey(authtypes.StoreKey)
	bankStoreKey := storetypes.NewKVStoreKey(banktypes.StoreKey)
	tkey := storetypes.NewTransientStoreKey("transient_test")

	// Setup context with all stores
	keys := map[string]*storetypes.KVStoreKey{
		poatypes.StoreKey:  storeKey,
		govtypes.StoreKey:  govStoreKey,
		authtypes.StoreKey: authStoreKey,
		banktypes.StoreKey: bankStoreKey,
	}
	ctx := testutil.DefaultContextWithKeys(keys, map[string]*storetypes.TransientStoreKey{"transient_test": tkey}, map[string]*storetypes.MemoryStoreKey{})

	// Set default consensus params with ed25519 as allowed validator pubkey type.
	ctx = ctx.WithConsensusParams(cmtproto.ConsensusParams{
		Validator: &cmtproto.ValidatorParams{
			PubKeyTypes: cmttypes.DefaultValidatorParams().PubKeyTypes,
		},
	})

	// Setup codec
	encCfg := moduletestutil.MakeTestEncodingConfig()
	std.RegisterInterfaces(encCfg.InterfaceRegistry)
	authtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	cdc := encCfg.Codec

	// Setup auth keeper
	authStoreService := runtime.NewKVStoreService(authStoreKey)
	maccPerms := map[string][]string{
		govtypes.ModuleName:        {},
		authtypes.FeeCollectorName: {authtypes.Minter},
		poatypes.ModuleName:        {},
	}
	authKeeper := authkeeper.NewAccountKeeper(
		cdc,
		authStoreService,
		authtypes.ProtoBaseAccount,
		maccPerms,
		addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	bankStoreService := runtime.NewKVStoreService(bankStoreKey)
	bankKeeper := bankkeeper.NewBaseKeeper(cdc, bankStoreService, authKeeper, BlockedAddresses(maccPerms), authtypes.NewModuleAddress(govtypes.ModuleName).String(), log.NewTestLogger(t))

	// Setup POA keeper
	poaStoreService := runtime.NewKVStoreService(storeKey)
	poaTransientStoreService := runtime.NewTransientStoreService(tkey)
	poaKeeper := NewKeeper(cdc, poaStoreService, poaTransientStoreService, authKeeper, bankKeeper)

	// Create the POA tally function
	tallyFn := NewPOACalculateVoteResultsAndVotingPowerFn(*poaKeeper)

	// Setup gov keeper with the POA tally function
	govStoreService := runtime.NewKVStoreService(govStoreKey)
	msgRouter := baseapp.NewMsgServiceRouter()
	govKeeper := govkeeper.NewKeeper(
		cdc,
		govStoreService,
		authKeeper,
		bankKeeper, // bank keeper
		nil,        // distribution keeper
		msgRouter,
		govtypes.DefaultConfig(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		tallyFn,
	)

	// Set the POA governance hooks to validate voters
	govKeeper.SetHooks(poaKeeper.NewGovHooks())

	// Initialize gov params
	err := govKeeper.Params.Set(ctx, govv1.DefaultParams())
	require.NoError(t, err)

	return &testFixture{
		ctx:        ctx,
		poaKeeper:  poaKeeper,
		govKeeper:  govKeeper,
		authKeeper: authKeeper,
		bankKeeper: bankKeeper,
		cdc:        cdc,
	}
}

// Helper to create a validator in POA keeper
func createValidator(t *testing.T, f *testFixture, index int, power int64) (string, sdk.ConsAddress) {
	t.Helper()

	// Create a proper test address for the operator
	operatorAddr := sdk.AccAddress(fmt.Sprintf("operator%d", index))
	operatorAddrStr := operatorAddr.String()

	// Create a proper consensus address
	consAddr := sdk.ConsAddress(fmt.Sprintf("cons%d", index))

	pubKey := ed25519.GenPrivKey().PubKey()

	validator := poatypes.Validator{
		PubKey: types.UnsafePackAny(pubKey),
		Power:  power,
		Metadata: &poatypes.ValidatorMetadata{
			Moniker:         fmt.Sprintf("validator-%d", index),
			OperatorAddress: operatorAddrStr,
		},
	}

	err := f.poaKeeper.CreateValidator(f.ctx, consAddr, validator, true)
	require.NoError(t, err)

	return operatorAddrStr, consAddr
}

// Helper to create a proposal and activate voting period
// proposerAddr must be an active validator address
func createProposal(t *testing.T, f *testFixture, proposerAddr sdk.AccAddress) uint64 {
	t.Helper()

	proposal, err := f.govKeeper.SubmitProposal(f.ctx, []sdk.Msg{}, "", "Test Proposal", "Description", proposerAddr, false)
	require.NoError(t, err)

	// Activate voting period
	err = f.govKeeper.ActivateVotingPeriod(f.ctx, proposal)
	require.NoError(t, err)

	return proposal.Id
}

// Helper to submit a vote directly to the gov keeper store (bypassing hooks for tally tests)
func submitVoteDirectly(t *testing.T, f *testFixture, proposalID uint64, voter string, options govv1.WeightedVoteOptions) {
	t.Helper()

	voterAddr, err := sdk.AccAddressFromBech32(voter)
	require.NoError(t, err)

	vote := govv1.Vote{
		ProposalId: proposalID,
		Voter:      voter,
		Options:    options,
	}

	key := collections.Join(proposalID, voterAddr)
	err = f.govKeeper.Votes.Set(f.ctx, key, vote)
	require.NoError(t, err)
}
