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

package integration_test

import (
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log/v2"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa"
	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/keeper"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type fixture struct {
	app *integration.App

	cdc       codec.Codec
	keys      map[string]*storetypes.KVStoreKey
	tkeys     map[string]*storetypes.TransientStoreKey
	poaKeeper *keeper.Keeper

	adminAddr sdk.AccAddress
	valPubKey ed25519.PubKey
}

func initFixture(tb testing.TB) *fixture {
	tb.Helper()

	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, poatypes.StoreKey,
	)
	tkeys := storetypes.NewTransientStoreKeys(poatypes.TransientStoreKey)

	cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{}, poa.AppModuleBasic{}).Codec

	logger := log.NewTestLogger(tb)
	cms := integration.CreateMultiStore(keys, logger)

	// Mount transient stores manually since integration framework doesn't support them
	for _, tkey := range tkeys {
		cms.MountStoreWithDB(tkey, storetypes.StoreTypeTransient, nil)
	}

	// Load the stores
	require.NoError(tb, cms.LoadLatestVersion())

	newCtx := sdk.NewContext(cms, cmtproto.Header{}, true, logger)

	authority := authtypes.NewModuleAddress("gov")

	maccPerms := map[string][]string{
		authtypes.FeeCollectorName: nil,
	}

	accountKeeper := authkeeper.NewAccountKeeper(
		cdc,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		address.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		authority.String(),
	)

	// Create fee collector module account
	feeCollectorAcc := authtypes.NewEmptyModuleAccount(authtypes.FeeCollectorName)
	accountKeeper.SetModuleAccount(newCtx, feeCollectorAcc)

	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		accountKeeper,
		map[string]bool{},
		authority.String(),
		log.NewNopLogger(),
	)

	poaKeeper := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[poatypes.StoreKey]),
		runtime.NewTransientStoreService(tkeys[poatypes.TransientStoreKey]),
		accountKeeper,
		bankKeeper,
	)

	authModule := auth.NewAppModule(cdc, accountKeeper, authsims.RandomGenesisAccounts)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper)
	poaModule := poa.NewAppModule(cdc, poaKeeper)

	adminAddr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	valPubKey := *ed25519.GenPrivKey().PubKey().(*ed25519.PubKey)

	// Mount transient stores in the BaseApp so they're properly managed across commits
	// Not doing this will fail the commit with a transient_poa key not found error.
	mountTransientStores := func(bapp *baseapp.BaseApp) {
		for _, tkey := range tkeys {
			bapp.MountStore(tkey, storetypes.StoreTypeTransient)
		}
	}

	integrationApp := integration.NewIntegrationApp(newCtx, logger, keys, cdc, map[string]appmodule.AppModule{
		authtypes.StoreKey: authModule,
		banktypes.StoreKey: bankModule,
		poatypes.StoreKey:  poaModule,
	}, mountTransientStores)

	return &fixture{
		app:       integrationApp,
		cdc:       cdc,
		keys:      keys,
		tkeys:     tkeys,
		poaKeeper: poaKeeper,
		adminAddr: adminAddr,
		valPubKey: valPubKey,
	}
}

func TestTransientStoreQueuesValidatorUpdates(t *testing.T) {
	f := initFixture(t)

	ctx := sdk.UnwrapSDKContext(f.app.Context())

	// Initialize POA module with admin
	params := poatypes.Params{
		Admin: f.adminAddr.String(),
	}
	require.NoError(t, f.poaKeeper.UpdateParams(ctx, params))

	// Create a validator with 0 power initially
	consAddr := sdk.ConsAddress(f.valPubKey.Address())
	pubKeyAny := codectypes.UnsafePackAny(&f.valPubKey)
	validator := poatypes.Validator{
		PubKey: pubKeyAny,
		Power:  0,
		Metadata: &poatypes.ValidatorMetadata{
			Moniker:         "test-validator",
			OperatorAddress: f.adminAddr.String(),
		},
	}

	require.NoError(t, f.poaKeeper.CreateValidator(ctx, consAddr, validator, true))

	// Update the validator power to trigger transient store writes
	updatedValidator := validator
	updatedValidator.Power = 2000
	err := f.poaKeeper.UpdateValidator(ctx, consAddr, updatedValidator)
	require.NoError(t, err)

	// Reap validator updates - should return the queued update from transient store
	queuedUpdates := f.poaKeeper.ReapValidatorUpdates(ctx)
	require.Len(t, queuedUpdates, 1, "should have exactly 1 queued update in transient store")
	assert.Equal(t, int64(2000), queuedUpdates[0].Power)

	// Verify that reading again in the same context returns the same data
	// (transient store persists within a block/context)
	queuedUpdates2 := f.poaKeeper.ReapValidatorUpdates(ctx)
	require.Len(t, queuedUpdates2, 1, "transient store persists within same context")

	// Verify EndBlocker returns the queued updates
	endBlockUpdates, err := f.poaKeeper.EndBlocker(ctx)
	require.NoError(t, err)
	require.Len(t, endBlockUpdates, 1, "EndBlocker should return queued updates")
	assert.Equal(t, int64(2000), endBlockUpdates[0].Power)

	// Now test that transient stores are cleared after block commit
	// Finalize and commit the current block
	_, err = f.app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: f.app.LastBlockHeight() + 1})
	require.NoError(t, err)
	_, err = f.app.Commit()
	require.NoError(t, err)

	// Get the BaseApp's multistore and create a new context for the next block
	// This simulates what happens in production when a new block begins
	cms := f.app.CommitMultiStore()
	newBlockCtx := sdk.NewContext(cms, cmtproto.Header{Height: ctx.BlockHeight() + 1}, false, ctx.Logger())

	// Query the transient store in the new block context - should be empty
	// because transient stores are cleared on commit
	queuedUpdatesAfterCommit := f.poaKeeper.ReapValidatorUpdates(newBlockCtx)
	assert.Empty(t, queuedUpdatesAfterCommit, "transient store should be cleared after block commit")
}
