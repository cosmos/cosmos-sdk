package staking_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func bootstrapGenesisTest(numAddrs int) (*simapp.SimApp, sdk.Context, []sdk.AccAddress) {
	_, app, ctx := getBaseSimappWithCustomKeeper()

	addrDels, _ := generateAddresses(app, ctx, numAddrs, sdk.NewInt(10000))
	return app, ctx, addrDels
}

func TestInitGenesis(t *testing.T) {
	app, ctx, addrs := bootstrapGenesisTest(10)

	valTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 1)

	params := app.StakingKeeper.GetParams(ctx)
	validators := app.StakingKeeper.GetAllValidators(ctx)
	var delegations []types.Delegation

	pk0, err := codectypes.NewAnyWithValue(PKs[0])
	require.NoError(t, err)

	pk1, err := codectypes.NewAnyWithValue(PKs[1])
	require.NoError(t, err)

	// initialize the validators
	bondedVal1 := types.Validator{
		OperatorAddress: sdk.ValAddress(addrs[0]).String(),
		ConsensusPubkey: pk0,
		Status:          types.Bonded,
		Tokens:          valTokens,
		DelegatorShares: valTokens.ToDec(),
		Description:     types.NewDescription("hoop", "", "", "", ""),
	}
	bondedVal2 := types.Validator{
		OperatorAddress: sdk.ValAddress(addrs[1]).String(),
		ConsensusPubkey: pk1,
		Status:          types.Bonded,
		Tokens:          valTokens,
		DelegatorShares: valTokens.ToDec(),
		Description:     types.NewDescription("bloop", "", "", "", ""),
	}

	// append new bonded validators to the list
	validators = append(validators, bondedVal1, bondedVal2)
	log.Printf("%#v", len(validators))
	// mint coins in the bonded pool representing the validators coins
	require.NoError(t,
		simapp.FundModuleAccount(
			app.BankKeeper,
			ctx,
			types.BondedPoolName,
			sdk.NewCoins(
				sdk.NewCoin(params.BondDenom, valTokens.MulRaw((int64)(len(validators)))),
			),
		),
	)
	genesisState := types.NewGenesisState(params, validators, delegations)
	vals := staking.InitGenesis(ctx, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, genesisState)

	actualGenesis := staking.ExportGenesis(ctx, app.StakingKeeper)
	require.Equal(t, genesisState.Params, actualGenesis.Params)
	require.Equal(t, genesisState.Delegations, actualGenesis.Delegations)
	require.EqualValues(t, app.StakingKeeper.GetAllValidators(ctx), actualGenesis.Validators)

	// Ensure validators have addresses.
	vals2, err := staking.WriteValidators(ctx, app.StakingKeeper)
	require.NoError(t, err)
	for _, val := range vals2 {
		require.NotEmpty(t, val.Address)
	}

	// now make sure the validators are bonded and intra-tx counters are correct
	resVal, found := app.StakingKeeper.GetValidator(ctx, sdk.ValAddress(addrs[0]))
	require.True(t, found)
	require.Equal(t, types.Bonded, resVal.Status)

	resVal, found = app.StakingKeeper.GetValidator(ctx, sdk.ValAddress(addrs[1]))
	require.True(t, found)
	require.Equal(t, types.Bonded, resVal.Status)

	abcivals := make([]abci.ValidatorUpdate, len(vals))
	for i, val := range validators {
		abcivals[i] = val.ABCIValidatorUpdate(app.StakingKeeper.PowerReduction(ctx))
	}

	require.Equal(t, abcivals, vals)
}

func TestInitGenesis_PoolsBalanceMismatch(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.NewContext(false, tmproto.Header{})

	consPub, err := codectypes.NewAnyWithValue(PKs[0])
	require.NoError(t, err)

	// create mock validator
	validator := types.Validator{
		OperatorAddress: sdk.ValAddress("12345678901234567890").String(),
		ConsensusPubkey: consPub,
		Jailed:          false,
		Tokens:          sdk.NewInt(10),
		DelegatorShares: sdk.NewInt(10).ToDec(),
		Description:     types.NewDescription("bloop", "", "", "", ""),
	}
	// valid params
	params := types.Params{
		UnbondingTime: 10000,
		MaxValidators: 1,
		MaxEntries:    10,
		BondDenom:     "stake",
	}

	// test

	require.Panics(t, func() {
		// setting validator status to bonded so the balance counts towards bonded pool
		validator.Status = types.Bonded
		staking.InitGenesis(ctx, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, &types.GenesisState{
			Params:     params,
			Validators: []types.Validator{validator},
		})
	}, "should panic because bonded pool balance is different from bonded pool coins")

	require.Panics(t, func() {
		// setting validator status to unbonded so the balance counts towards not bonded pool
		validator.Status = types.Unbonded
		staking.InitGenesis(ctx, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, &types.GenesisState{
			Params:     params,
			Validators: []types.Validator{validator},
		})
	}, "should panic because not bonded pool balance is different from not bonded pool coins")
}

func TestInitGenesisLargeValidatorSet(t *testing.T) {
	size := 200
	require.True(t, size > 100)

	app, ctx, addrs := bootstrapGenesisTest(200)

	params := app.StakingKeeper.GetParams(ctx)
	delegations := []types.Delegation{}
	validators := make([]types.Validator, size)
	var err error

	bondedPoolAmt := sdk.ZeroInt()
	for i := range validators {
		validators[i], err = types.NewValidator(sdk.ValAddress(addrs[i]),
			PKs[i], types.NewDescription(fmt.Sprintf("#%d", i), "", "", "", ""))
		require.NoError(t, err)
		validators[i].Status = types.Bonded

		tokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 1)
		if i < 100 {
			tokens = app.StakingKeeper.TokensFromConsensusPower(ctx, 2)
		}
		validators[i].Tokens = tokens
		validators[i].DelegatorShares = tokens.ToDec()
		// add bonded coins
		bondedPoolAmt = bondedPoolAmt.Add(tokens)
	}

	genesisState := types.NewGenesisState(params, validators, delegations)

	// mint coins in the bonded pool representing the validators coins
	require.NoError(t,
		simapp.FundModuleAccount(
			app.BankKeeper,
			ctx,
			types.BondedPoolName,
			sdk.NewCoins(sdk.NewCoin(params.BondDenom, bondedPoolAmt)),
		),
	)

	vals := staking.InitGenesis(ctx, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, genesisState)

	abcivals := make([]abci.ValidatorUpdate, 100)
	for i, val := range validators[:100] {
		abcivals[i] = val.ABCIValidatorUpdate(app.StakingKeeper.PowerReduction(ctx))
	}

	require.Equal(t, abcivals, vals)
}

func TestValidateGenesis(t *testing.T) {
	genValidators1 := make([]types.Validator, 1, 5)
	pk := ed25519.GenPrivKey().PubKey()
	genValidators1[0] = teststaking.NewValidator(t, sdk.ValAddress(pk.Address()), pk)
	genValidators1[0].Tokens = sdk.OneInt()
	genValidators1[0].DelegatorShares = sdk.OneDec()

	tests := []struct {
		name    string
		mutate  func(*types.GenesisState)
		wantErr bool
	}{
		{"default", func(*types.GenesisState) {}, false},
		// validate genesis validators
		{"duplicate validator", func(data *types.GenesisState) {
			data.Validators = genValidators1
			data.Validators = append(data.Validators, genValidators1[0])
		}, true},
		{"no delegator shares", func(data *types.GenesisState) {
			data.Validators = genValidators1
			data.Validators[0].DelegatorShares = sdk.ZeroDec()
		}, true},
		{"jailed and bonded validator", func(data *types.GenesisState) {
			data.Validators = genValidators1
			data.Validators[0].Jailed = true
			data.Validators[0].Status = types.Bonded
		}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			genesisState := types.DefaultGenesisState()
			tt.mutate(genesisState)
			if tt.wantErr {
				assert.Error(t, staking.ValidateGenesis(genesisState))
			} else {
				assert.NoError(t, staking.ValidateGenesis(genesisState))
			}
		})
	}
}
