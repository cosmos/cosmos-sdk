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

package testutil

import (
	"encoding/json"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log/v2"
	sdkmath "cosmossdk.io/math"

	sdkapp "github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	group "github.com/cosmos/cosmos-sdk/enterprise/group/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/enterprise/group/x/group/keeper"
	groupmodule "github.com/cosmos/cosmos-sdk/enterprise/group/x/group/module"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type groupAppModule struct {
	groupmodule.AppModule
	storeKey *storetypes.KVStoreKey
}

func (m groupAppModule) StoreKeys() map[string]*storetypes.KVStoreKey {
	return map[string]*storetypes.KVStoreKey{group.StoreKey: m.storeKey}
}

func (groupAppModule) ModuleAccountPermissions() map[string][]string {
	return map[string][]string{}
}

// SetupApp creates an SDKApp with the group module registered and a single
// validator genesis, returning the app and the group keeper.
func SetupApp(t *testing.T) (*sdkapp.SDKApp, groupkeeper.Keeper) {
	t.Helper()

	opts := simtestutil.AppOptionsMap{
		flags.FlagHome:    t.TempDir(),
		flags.FlagChainID: "test-chain",
	}
	cfg := sdkapp.DefaultSDKAppConfig("app", opts)
	ta := sdkapp.NewSDKApp(log.NewNopLogger(), dbm.NewMemDB(), nil, cfg)

	groupStoreKey := storetypes.NewKVStoreKey(group.StoreKey)
	groupConfig := group.DefaultConfig()
	k := groupkeeper.NewKeeper(
		groupStoreKey,
		ta.AppCodec(),
		ta.MsgServiceRouter(),
		ta.AccountKeeper,
		groupConfig,
	)
	mod := groupAppModule{
		AppModule: groupmodule.NewAppModule(ta.AppCodec(), k, ta.AccountKeeper, ta.BankKeeper, ta.InterfaceRegistry()),
		storeKey:  groupStoreKey,
	}
	if err := ta.AddModules(mod); err != nil {
		t.Fatalf("failed to add group module: %v", err)
	}

	ta.LoadModules()

	if err := ta.LoadLatestVersion(); err != nil {
		t.Fatalf("failed to load latest version: %v", err)
	}

	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	if err != nil {
		t.Fatalf("failed to get pub key: %v", err)
	}
	validator := cmttypes.NewValidator(pubKey, 1)
	valSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{validator})

	priv := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(priv.PubKey().Address().Bytes(), priv.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100000000000000))),
	}

	genesisState, err := buildGroupGenesis(ta.AppCodec(), ta.DefaultGenesis(), valSet, []authtypes.GenesisAccount{acc}, balance)
	if err != nil {
		t.Fatalf("failed to build genesis: %v", err)
	}

	stateBytes, err := cmtjson.MarshalIndent(genesisState, "", " ")
	if err != nil {
		t.Fatalf("failed to marshal genesis: %v", err)
	}

	if _, err := ta.InitChain(&abci.RequestInitChain{
		ChainId:         "test-chain",
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: simtestutil.DefaultConsensusParams,
		AppStateBytes:   stateBytes,
	}); err != nil {
		t.Fatalf("failed to init chain: %v", err)
	}

	if _, err := ta.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             ta.LastBlockHeight() + 1,
		NextValidatorsHash: valSet.Hash(),
	}); err != nil {
		t.Fatalf("failed to finalize block: %v", err)
	}

	return ta, k
}

func buildGroupGenesis(
	cdc codec.Codec,
	genesisState map[string]json.RawMessage,
	valSet *cmttypes.ValidatorSet,
	genAccs []authtypes.GenesisAccount,
	balances ...banktypes.Balance,
) (map[string]json.RawMessage, error) {
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = cdc.MustMarshalJSON(authGenesis)

	bondAmt := sdk.DefaultPowerReduction
	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))

	for _, val := range valSet.Validators {
		pk, err := cryptocodec.FromCmtPubKeyInterface(val.PubKey)
		if err != nil {
			return nil, err
		}
		pkAny, err := codectypes.NewAnyWithValue(pk)
		if err != nil {
			return nil, err
		}
		validators = append(validators, stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   sdkmath.LegacyOneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec()),
			MinSelfDelegation: sdkmath.ZeroInt(),
		})
		delegations = append(delegations, stakingtypes.NewDelegation(
			genAccs[0].GetAddress().String(),
			sdk.ValAddress(val.Address).String(),
			sdkmath.LegacyOneDec(),
		))
	}

	stakingGenesis := stakingtypes.NewGenesisState(stakingtypes.DefaultParams(), validators, delegations)
	genesisState[stakingtypes.ModuleName] = cdc.MustMarshalJSON(stakingGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		totalSupply = totalSupply.Add(b.Coins...)
	}
	for range delegations {
		totalSupply = totalSupply.Add(sdk.NewCoin(sdk.DefaultBondDenom, bondAmt))
	}
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, bondAmt)},
	})
	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply, []banktypes.Metadata{}, []banktypes.SendEnabled{})
	genesisState[banktypes.ModuleName] = cdc.MustMarshalJSON(bankGenesis)

	return genesisState, nil
}
