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
	"testing"

	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"

	"cosmossdk.io/log/v2"
	"cosmossdk.io/math"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/std"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type anteTestFixture struct {
	ctx        sdk.Context
	authKeeper authkeeper.AccountKeeper
	bankKeeper bankkeeper.BaseKeeper
}

func setupAnteTest(t *testing.T) *anteTestFixture {
	t.Helper()

	authStoreKey := storetypes.NewKVStoreKey(authtypes.StoreKey)
	bankStoreKey := storetypes.NewKVStoreKey(banktypes.StoreKey)

	ctx := testutil.DefaultContextWithKeys(
		map[string]*storetypes.KVStoreKey{
			authtypes.StoreKey: authStoreKey,
			banktypes.StoreKey: bankStoreKey,
		},
		map[string]*storetypes.TransientStoreKey{},
		map[string]*storetypes.MemoryStoreKey{},
	)
	ctx = ctx.WithBlockHeight(1)

	encCfg := moduletestutil.MakeTestEncodingConfig()
	std.RegisterInterfaces(encCfg.InterfaceRegistry)
	authtypes.RegisterInterfaces(encCfg.InterfaceRegistry)
	cdc := encCfg.Codec

	maccPerms := map[string][]string{
		authtypes.FeeCollectorName: {authtypes.Minter},
		poatypes.ModuleName:        {authtypes.Minter},
	}

	authKeeper := authkeeper.NewAccountKeeper(
		cdc,
		runtime.NewKVStoreService(authStoreKey),
		authtypes.ProtoBaseAccount,
		maccPerms,
		addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		authtypes.NewModuleAddress("gov").String(),
	)

	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		runtime.NewKVStoreService(bankStoreKey),
		authKeeper,
		map[string]bool{},
		authtypes.NewModuleAddress("gov").String(),
		log.NewTestLogger(t),
	)

	return &anteTestFixture{
		ctx:        ctx,
		authKeeper: authKeeper,
		bankKeeper: bankKeeper,
	}
}

func (f *anteTestFixture) fundAccount(t *testing.T, addr sdk.AccAddress, amount sdk.Coins) {
	t.Helper()
	err := f.bankKeeper.MintCoins(f.ctx, poatypes.ModuleName, amount)
	require.NoError(t, err)
	err = f.bankKeeper.SendCoinsFromModuleToAccount(f.ctx, poatypes.ModuleName, addr, amount)
	require.NoError(t, err)
}

func (f *anteTestFixture) createAccount(t *testing.T, addr sdk.AccAddress) {
	t.Helper()
	acc := f.authKeeper.NewAccountWithAddress(f.ctx, addr)
	f.authKeeper.SetAccount(f.ctx, acc)
}

func (f *anteTestFixture) assertFeeCollectorEmpty(t *testing.T) {
	t.Helper()
	addr := f.authKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	bal := f.bankKeeper.GetAllBalances(f.ctx, addr)
	require.True(t, bal.IsZero(), "fee_collector should remain empty, got %s", bal)
}

func (f *anteTestFixture) poaModuleBalance(t *testing.T) sdk.Coins {
	t.Helper()
	addr := f.authKeeper.GetModuleAddress(poatypes.ModuleName)
	return f.bankKeeper.GetAllBalances(f.ctx, addr)
}

// TestPOADeductFeeDecorator_ZeroGas mirrors TestDeductFeeDecorator_ZeroGas
// from x/auth/ante/fee_test.go.
func TestPOADeductFeeDecorator_ZeroGas(t *testing.T) {
	tests := []struct {
		name     string
		simulate bool
		expErr   bool
	}{
		{"zero gas rejected in non-simulate", false, true},
		{"zero gas accepted in simulate", true, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := setupAnteTest(t)

			userAddr := sdk.AccAddress("testuser")
			f.createAccount(t, userAddr)
			f.fundAccount(t, userAddr, sdk.NewCoins(sdk.NewInt64Coin("stake", 100)))

			dfd := NewPOADeductFeeDecorator(f.authKeeper, f.bankKeeper, nil, nil)
			anteHandler := sdk.ChainAnteDecorators(dfd)

			tx := mockFeeTx{fee: sdk.NewCoins(sdk.NewInt64Coin("stake", 50)), gas: 0, feePayer: userAddr}

			_, err := anteHandler(f.ctx, tx, tc.simulate)
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			f.assertFeeCollectorEmpty(t)
		})
	}
}

// TestPOAEnsureMempoolFees mirrors TestEnsureMempoolFees from
// x/auth/ante/fee_test.go.
func TestPOAEnsureMempoolFees(t *testing.T) {
	tests := []struct {
		name        string
		minGasPrice math.LegacyDec
		isCheckTx   bool
		simulate    bool
		expErr      bool
		expPriority int64
	}{
		{
			name:        "high min gas price in CheckTx rejects insufficient fee",
			minGasPrice: math.LegacyNewDec(20),
			isCheckTx:   true,
			simulate:    false,
			expErr:      true,
		},
		{
			name:        "simulate bypasses min gas price check",
			minGasPrice: math.LegacyNewDec(20),
			isCheckTx:   true,
			simulate:    true,
			expErr:      false,
		},
		{
			name:        "DeliverTx bypasses min gas price check",
			minGasPrice: math.LegacyNewDec(20),
			isCheckTx:   false,
			simulate:    false,
			expErr:      false,
		},
		{
			name:        "low min gas price in CheckTx passes with correct priority",
			minGasPrice: math.LegacyNewDec(1).Quo(math.LegacyNewDec(100000)),
			isCheckTx:   true,
			simulate:    false,
			expErr:      false,
			expPriority: 10, // 150atom / 15gas = 10
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := setupAnteTest(t)

			userAddr := sdk.AccAddress("testuser")
			f.createAccount(t, userAddr)
			f.fundAccount(t, userAddr, sdk.NewCoins(sdk.NewInt64Coin("atom", 1000)))

			dfd := NewPOADeductFeeDecorator(f.authKeeper, f.bankKeeper, nil, nil)
			anteHandler := sdk.ChainAnteDecorators(dfd)

			feeAmount := sdk.NewCoins(sdk.NewInt64Coin("atom", 150))
			tx := mockFeeTx{fee: feeAmount, gas: 15, feePayer: userAddr}

			gasPrices := []sdk.DecCoin{sdk.NewDecCoinFromDec("atom", tc.minGasPrice)}
			ctx := f.ctx.WithMinGasPrices(gasPrices).WithIsCheckTx(tc.isCheckTx)

			newCtx, err := anteHandler(ctx, tx, tc.simulate)
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tc.expPriority != 0 {
				require.Equal(t, tc.expPriority, newCtx.Priority())
			}

			f.assertFeeCollectorEmpty(t)

			if !tc.expErr && !tc.simulate {
				require.False(t, f.poaModuleBalance(t).IsZero(), "poa module should have received fees")
			}
		})
	}
}

// TestPOADeductFees mirrors TestDeductFees from x/auth/ante/fee_test.go.
func TestPOADeductFees(t *testing.T) {
	feeAmount := sdk.NewCoins(sdk.NewInt64Coin("stake", 150))

	tests := []struct {
		name       string
		fundUser   sdk.Coins
		expErr     bool
		expPOABal  sdk.Coins
		expUserBal sdk.Coins
	}{
		{
			name:     "insufficient funds errors",
			fundUser: nil,
			expErr:   true,
		},
		{
			name:       "funded account succeeds and fees go to poa module",
			fundUser:   sdk.NewCoins(sdk.NewInt64Coin("stake", 200)),
			expErr:     false,
			expPOABal:  feeAmount,
			expUserBal: sdk.NewCoins(sdk.NewInt64Coin("stake", 50)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := setupAnteTest(t)

			userAddr := sdk.AccAddress("testuser")
			f.createAccount(t, userAddr)
			if tc.fundUser != nil {
				f.fundAccount(t, userAddr, tc.fundUser)
			}

			dfd := NewPOADeductFeeDecorator(
				f.authKeeper, f.bankKeeper, nil,
				func(ctx sdk.Context, tx sdk.Tx) (sdk.Coins, int64, error) {
					return feeAmount, 0, nil
				},
			)
			anteHandler := sdk.ChainAnteDecorators(dfd)

			tx := mockFeeTx{fee: feeAmount, gas: 200000, feePayer: userAddr}

			_, err := anteHandler(f.ctx, tx, false)
			if tc.expErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tc.expPOABal, f.poaModuleBalance(t))
			f.assertFeeCollectorEmpty(t)

			userBal := f.bankKeeper.GetAllBalances(f.ctx, userAddr)
			require.Equal(t, tc.expUserBal, userBal)
		})
	}
}

// mockFeeTx implements sdk.Tx and sdk.FeeTx for testing.
type mockFeeTx struct {
	fee      sdk.Coins
	gas      uint64
	feePayer sdk.AccAddress
}

func (m mockFeeTx) GetMsgs() []sdk.Msg                    { return nil }
func (m mockFeeTx) GetMsgsV2() ([]protov2.Message, error) { return nil, nil }
func (m mockFeeTx) ValidateBasic() error                   { return nil }
func (m mockFeeTx) GetGas() uint64                         { return m.gas }
func (m mockFeeTx) GetFee() sdk.Coins                      { return m.fee }
func (m mockFeeTx) FeePayer() []byte                       { return m.feePayer }
func (m mockFeeTx) FeeGranter() []byte                     { return nil }
