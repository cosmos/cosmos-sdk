package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func bootstrapGenesisTest(t *testing.T, numAddrs int) (*simapp.SimApp, sdk.Context, []sdk.AccAddress) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addrDels, _ := generateAddresses(app, ctx, numAddrs)
	return app, ctx, addrDels
}

func TestInitGenesis(t *testing.T) {
	app, ctx, addrs := bootstrapGenesisTest(t, 10)

	valTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 1)

	params := app.StakingKeeper.GetParams(ctx)
	validators := app.StakingKeeper.GetAllValidators(ctx)
	require.Len(t, validators, 1)
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
		DelegatorShares: sdk.NewDecFromInt(valTokens),
		Description:     types.NewDescription("hoop", "", "", "", ""),
	}
	bondedVal2 := types.Validator{
		OperatorAddress: sdk.ValAddress(addrs[1]).String(),
		ConsensusPubkey: pk1,
		Status:          types.Bonded,
		Tokens:          valTokens,
		DelegatorShares: sdk.NewDecFromInt(valTokens),
		Description:     types.NewDescription("bloop", "", "", "", ""),
	}

	// append new bonded validators to the list
	validators = append(validators, bondedVal1, bondedVal2)

	// mint coins in the bonded pool representing the validators coins
	i2 := len(validators) - 1 // -1 to exclude genesis validator
	require.NoError(t,
		testutil.FundModuleAccount(
			app.BankKeeper,
			ctx,
			types.BondedPoolName,
			sdk.NewCoins(
				sdk.NewCoin(params.BondDenom, valTokens.MulRaw((int64)(i2))),
			),
		),
	)

	genesisDelegations := app.StakingKeeper.GetAllDelegations(ctx)
	delegations = append(delegations, genesisDelegations...)

	genesisState := types.NewGenesisState(params, validators, delegations)
	vals := app.StakingKeeper.InitGenesis(ctx, genesisState)

	actualGenesis := app.StakingKeeper.ExportGenesis(ctx)
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

	validators = validators[1:] // remove genesis validator
	for i, val := range validators {
		abcivals[i] = val.ABCIValidatorUpdate(app.StakingKeeper.PowerReduction(ctx))
	}

	require.Equal(t, abcivals, vals)
}

func TestInitGenesis_PoolsBalanceMismatch(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.NewContext(false, tmproto.Header{})

	consPub, err := codectypes.NewAnyWithValue(PKs[0])
	require.NoError(t, err)

	validator := types.Validator{
		OperatorAddress: sdk.ValAddress("12345678901234567890").String(),
		ConsensusPubkey: consPub,
		Jailed:          false,
		Tokens:          sdk.NewInt(10),
		DelegatorShares: sdk.NewDecFromInt(sdk.NewInt(10)),
		Description:     types.NewDescription("bloop", "", "", "", ""),
	}

	params := types.Params{
		UnbondingTime: 10000,
		MaxValidators: 1,
		MaxEntries:    10,
		BondDenom:     "stake",
	}

	require.Panics(t, func() {
		// setting validator status to bonded so the balance counts towards bonded pool
		validator.Status = types.Bonded
		app.StakingKeeper.InitGenesis(ctx, &types.GenesisState{
			Params:     params,
			Validators: []types.Validator{validator},
		})
	},
		"should panic because bonded pool balance is different from bonded pool coins",
	)

	require.Panics(t, func() {
		// setting validator status to unbonded so the balance counts towards not bonded pool
		validator.Status = types.Unbonded
		app.StakingKeeper.InitGenesis(ctx, &types.GenesisState{
			Params:     params,
			Validators: []types.Validator{validator},
		})
	},
		"should panic because not bonded pool balance is different from not bonded pool coins",
	)
}

func TestInitGenesisLargeValidatorSet(t *testing.T) {
	size := 200
	require.True(t, size > 100)

	app, ctx, addrs := bootstrapGenesisTest(t, 200)
	genesisValidators := app.StakingKeeper.GetAllValidators(ctx)

	params := app.StakingKeeper.GetParams(ctx)
	delegations := []types.Delegation{}
	validators := make([]types.Validator, size)

	var err error

	bondedPoolAmt := sdk.ZeroInt()
	for i := range validators {
		validators[i], err = types.NewValidator(
			sdk.ValAddress(addrs[i]),
			PKs[i],
			types.NewDescription(fmt.Sprintf("#%d", i), "", "", "", ""),
		)
		require.NoError(t, err)
		validators[i].Status = types.Bonded

		tokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 1)
		if i < 100 {
			tokens = app.StakingKeeper.TokensFromConsensusPower(ctx, 2)
		}

		validators[i].Tokens = tokens
		validators[i].DelegatorShares = sdk.NewDecFromInt(tokens)

		// add bonded coins
		bondedPoolAmt = bondedPoolAmt.Add(tokens)
	}

	validators = append(validators, genesisValidators...)
	genesisState := types.NewGenesisState(params, validators, delegations)

	// mint coins in the bonded pool representing the validators coins
	require.NoError(t,
		testutil.FundModuleAccount(
			app.BankKeeper,
			ctx,
			types.BondedPoolName,
			sdk.NewCoins(sdk.NewCoin(params.BondDenom, bondedPoolAmt)),
		),
	)

	vals := app.StakingKeeper.InitGenesis(ctx, genesisState)

	abcivals := make([]abci.ValidatorUpdate, 100)
	for i, val := range validators[:100] {
		abcivals[i] = val.ABCIValidatorUpdate(app.StakingKeeper.PowerReduction(ctx))
	}

	// remove genesis validator
	vals = vals[:100]
	require.Equal(t, abcivals, vals)
}
