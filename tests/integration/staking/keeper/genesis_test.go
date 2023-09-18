package keeper_test

import (
	"fmt"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func bootstrapGenesisTest(t *testing.T, numAddrs int) (*fixture, []sdk.AccAddress) {
	t.Helper()
	t.Parallel()
	f := initFixture(t)

	addrDels, _ := generateAddresses(f, numAddrs)
	return f, addrDels
}

func TestInitGenesis(t *testing.T) {
	f, addrs := bootstrapGenesisTest(t, 10)

	valTokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 1)

	pk0, err := codectypes.NewAnyWithValue(PKs[0])
	assert.NilError(t, err)

	bondedVal := types.Validator{
		OperatorAddress: sdk.ValAddress(addrs[0]).String(),
		ConsensusPubkey: pk0,
		Status:          types.Bonded,
		Tokens:          valTokens,
		DelegatorShares: math.LegacyNewDecFromInt(valTokens),
		Description:     types.NewDescription("hoop", "", "", "", ""),
	}
	assert.NilError(t, f.stakingKeeper.SetValidator(f.sdkCtx, bondedVal))

	params, err := f.stakingKeeper.Params.Get(f.sdkCtx)
	assert.NilError(t, err)

	validators, err := f.stakingKeeper.GetAllValidators(f.sdkCtx)
	assert.NilError(t, err)

	assert.Assert(t, len(validators) == 1)
	var delegations []types.Delegation

	pk1, err := codectypes.NewAnyWithValue(PKs[1])
	assert.NilError(t, err)

	pk2, err := codectypes.NewAnyWithValue(PKs[2])
	assert.NilError(t, err)

	// initialize the validators
	bondedVal1 := types.Validator{
		OperatorAddress: sdk.ValAddress(addrs[1]).String(),
		ConsensusPubkey: pk1,
		Status:          types.Bonded,
		Tokens:          valTokens,
		DelegatorShares: math.LegacyNewDecFromInt(valTokens),
		Description:     types.NewDescription("hoop", "", "", "", ""),
	}
	bondedVal2 := types.Validator{
		OperatorAddress: sdk.ValAddress(addrs[2]).String(),
		ConsensusPubkey: pk2,
		Status:          types.Bonded,
		Tokens:          valTokens,
		DelegatorShares: math.LegacyNewDecFromInt(valTokens),
		Description:     types.NewDescription("bloop", "", "", "", ""),
	}

	// append new bonded validators to the list
	validators = append(validators, bondedVal1, bondedVal2)

	// mint coins in the bonded pool representing the validators coins
	i2 := len(validators)
	assert.NilError(t,
		banktestutil.FundModuleAccount(
			f.sdkCtx,
			f.bankKeeper,
			types.BondedPoolName,
			sdk.NewCoins(
				sdk.NewCoin(params.BondDenom, valTokens.MulRaw((int64)(i2))),
			),
		),
	)

	genesisDelegations, err := f.stakingKeeper.GetAllDelegations(f.sdkCtx)
	assert.NilError(t, err)
	delegations = append(delegations, genesisDelegations...)

	genesisState := types.NewGenesisState(params, validators, delegations)
	vals := (f.stakingKeeper.InitGenesis(f.sdkCtx, genesisState))

	actualGenesis := (f.stakingKeeper.ExportGenesis(f.sdkCtx))
	assert.DeepEqual(t, genesisState.Params, actualGenesis.Params)
	assert.DeepEqual(t, genesisState.Delegations, actualGenesis.Delegations)

	allvals, err := f.stakingKeeper.GetAllValidators(f.sdkCtx)
	assert.NilError(t, err)
	assert.DeepEqual(t, allvals, actualGenesis.Validators)

	// Ensure validators have addresses.
	vals2, err := staking.WriteValidators(f.sdkCtx, (f.stakingKeeper))
	assert.NilError(t, err)

	for _, val := range vals2 {
		assert.Assert(t, val.Address.String() != "")
	}

	// now make sure the validators are bonded and intra-tx counters are correct
	resVal, found := (f.stakingKeeper.GetValidator(f.sdkCtx, sdk.ValAddress(addrs[1])))
	assert.Assert(t, found)
	assert.Equal(t, types.Bonded, resVal.Status)

	resVal, found = (f.stakingKeeper.GetValidator(f.sdkCtx, sdk.ValAddress(addrs[2])))
	assert.Assert(t, found)
	assert.Equal(t, types.Bonded, resVal.Status)

	abcivals := make([]abci.ValidatorUpdate, len(vals))

	for i, val := range validators {
		abcivals[i] = val.ABCIValidatorUpdate((f.stakingKeeper.PowerReduction(f.sdkCtx)))
	}

	assert.DeepEqual(t, abcivals, vals)
}

func TestInitGenesis_PoolsBalanceMismatch(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	consPub, err := codectypes.NewAnyWithValue(PKs[0])
	assert.NilError(t, err)

	validator := types.Validator{
		OperatorAddress: sdk.ValAddress("12345678901234567890").String(),
		ConsensusPubkey: consPub,
		Jailed:          false,
		Tokens:          math.NewInt(10),
		DelegatorShares: math.LegacyNewDecFromInt(math.NewInt(10)),
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
		f.stakingKeeper.InitGenesis(f.sdkCtx, &types.GenesisState{
			Params:     params,
			Validators: []types.Validator{validator},
		})
	},
	// "should panic because bonded pool balance is different from bonded pool coins",
	)

	require.Panics(t, func() {
		// setting validator status to unbonded so the balance counts towards not bonded pool
		validator.Status = types.Unbonded
		f.stakingKeeper.InitGenesis(f.sdkCtx, &types.GenesisState{
			Params:     params,
			Validators: []types.Validator{validator},
		})
	},
	// "should panic because not bonded pool balance is different from not bonded pool coins",
	)
}

func TestInitGenesisLargeValidatorSet(t *testing.T) {
	size := 200
	assert.Assert(t, size > 100)

	f, addrs := bootstrapGenesisTest(t, 200)
	genesisValidators, err := f.stakingKeeper.GetAllValidators(f.sdkCtx)
	assert.NilError(t, err)

	params, err := f.stakingKeeper.Params.Get(f.sdkCtx)
	assert.NilError(t, err)
	delegations := []types.Delegation{}
	validators := make([]types.Validator, size)

	bondedPoolAmt := math.ZeroInt()
	for i := range validators {
		validators[i], err = types.NewValidator(
			sdk.ValAddress(addrs[i]).String(),
			PKs[i],
			types.NewDescription(fmt.Sprintf("#%d", i), "", "", "", ""),
		)
		assert.NilError(t, err)
		validators[i].Status = types.Bonded

		tokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 1)
		if i < 100 {
			tokens = f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 2)
		}

		validators[i].Tokens = tokens
		validators[i].DelegatorShares = math.LegacyNewDecFromInt(tokens)

		// add bonded coins
		bondedPoolAmt = bondedPoolAmt.Add(tokens)
	}

	validators = append(validators, genesisValidators...)
	genesisState := types.NewGenesisState(params, validators, delegations)

	// mint coins in the bonded pool representing the validators coins
	assert.NilError(t,
		banktestutil.FundModuleAccount(
			f.sdkCtx,
			f.bankKeeper,
			types.BondedPoolName,
			sdk.NewCoins(sdk.NewCoin(params.BondDenom, bondedPoolAmt)),
		),
	)

	vals := f.stakingKeeper.InitGenesis(f.sdkCtx, genesisState)

	abcivals := make([]abci.ValidatorUpdate, 100)
	for i, val := range validators[:100] {
		abcivals[i] = val.ABCIValidatorUpdate(f.stakingKeeper.PowerReduction(f.sdkCtx))
	}

	// remove genesis validator
	vals = vals[:100]
	assert.DeepEqual(t, abcivals, vals)
}
