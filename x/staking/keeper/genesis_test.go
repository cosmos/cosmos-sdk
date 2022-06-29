package keeper_test

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (suite *KeeperTestSuite) TestInitGenesis() {
	addrs := simtestutil.AddTestAddrsIncremental(suite.bankKeeper, suite.stakingKeeper, suite.ctx, 10, sdk.NewInt(10000))

	valTokens := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 1)

	params := suite.stakingKeeper.GetParams(suite.ctx)
	validators := suite.stakingKeeper.GetAllValidators(suite.ctx)
	suite.Require().Len(validators, 1)
	var delegations []types.Delegation

	pk0, err := codectypes.NewAnyWithValue(PKs[0])
	suite.Require().NoError(err)

	pk1, err := codectypes.NewAnyWithValue(PKs[1])
	suite.Require().NoError(err)

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
	suite.Require().NoError(banktestutil.FundModuleAccount(
		suite.bankKeeper,
		suite.ctx,
		types.BondedPoolName,
		sdk.NewCoins(
			sdk.NewCoin(params.BondDenom, valTokens.MulRaw((int64)(i2))),
		),
	),
	)

	genesisDelegations := suite.stakingKeeper.GetAllDelegations(suite.ctx)
	delegations = append(delegations, genesisDelegations...)

	genesisState := types.NewGenesisState(params, validators, delegations)
	vals := suite.stakingKeeper.InitGenesis(suite.ctx, genesisState)

	actualGenesis := suite.stakingKeeper.ExportGenesis(suite.ctx)
	suite.Require().Equal(genesisState.Params, actualGenesis.Params)
	suite.Require().Equal(genesisState.Delegations, actualGenesis.Delegations)
	suite.Require().EqualValues(suite.stakingKeeper.GetAllValidators(suite.ctx), actualGenesis.Validators)

	// Ensure validators have addresses.
	vals2, err := staking.WriteValidators(suite.ctx, suite.stakingKeeper)
	suite.Require().NoError(err)

	for _, val := range vals2 {
		suite.Require().NotEmpty(val.Address)
	}

	// now make sure the validators are bonded and intra-tx counters are correct
	resVal, found := suite.stakingKeeper.GetValidator(suite.ctx, sdk.ValAddress(addrs[0]))
	suite.Require().True(found)
	suite.Require().Equal(types.Bonded, resVal.Status)

	resVal, found = suite.stakingKeeper.GetValidator(suite.ctx, sdk.ValAddress(addrs[1]))
	suite.Require().True(found)
	suite.Require().Equal(types.Bonded, resVal.Status)

	abcivals := make([]abci.ValidatorUpdate, len(vals))

	validators = validators[1:] // remove genesis validator
	for i, val := range validators {
		abcivals[i] = val.ABCIValidatorUpdate(suite.stakingKeeper.PowerReduction(suite.ctx))
	}

	suite.Require().Equal(abcivals, vals)
}

func (suite *KeeperTestSuite) TestInitGenesis_PoolsBalanceMismatch() {
	consPub, err := codectypes.NewAnyWithValue(PKs[0])
	suite.Require().NoError(err)

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

	suite.Require().Panics(func() {
		// setting validator status to bonded so the balance counts towards bonded pool
		validator.Status = types.Bonded
		suite.stakingKeeper.InitGenesis(suite.ctx, &types.GenesisState{
			Params:     params,
			Validators: []types.Validator{validator},
		})
	},
		"should panic because bonded pool balance is different from bonded pool coins",
	)

	suite.Require().Panics(func() {
		// setting validator status to unbonded so the balance counts towards not bonded pool
		validator.Status = types.Unbonded
		suite.stakingKeeper.InitGenesis(suite.ctx, &types.GenesisState{
			Params:     params,
			Validators: []types.Validator{validator},
		})
	},
		"should panic because not bonded pool balance is different from not bonded pool coins",
	)
}

func (suite *KeeperTestSuite) TestInitGenesisLargeValidatorSet() {
	size := 200

	addrDels := simtestutil.AddTestAddrsIncremental(suite.bankKeeper, suite.stakingKeeper, suite.ctx, size, sdk.NewInt(10000))
	addrs := simtestutil.ConvertAddrsToValAddrs(addrDels)
	genesisValidators := suite.stakingKeeper.GetAllValidators(suite.ctx)

	params := suite.stakingKeeper.GetParams(suite.ctx)
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
		suite.Require().NoError(err)
		validators[i].Status = types.Bonded

		tokens := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 1)
		if i < 100 {
			tokens = suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 2)
		}

		validators[i].Tokens = tokens
		validators[i].DelegatorShares = sdk.NewDecFromInt(tokens)

		// add bonded coins
		bondedPoolAmt = bondedPoolAmt.Add(tokens)
	}

	validators = append(validators, genesisValidators...)
	genesisState := types.NewGenesisState(params, validators, delegations)

	// mint coins in the bonded pool representing the validators coins
	suite.Require().NoError(banktestutil.FundModuleAccount(
		suite.bankKeeper,
		suite.ctx,
		types.BondedPoolName,
		sdk.NewCoins(sdk.NewCoin(params.BondDenom, bondedPoolAmt)),
	),
	)

	vals := suite.stakingKeeper.InitGenesis(suite.ctx, genesisState)

	abcivals := make([]abci.ValidatorUpdate, 100)
	for i, val := range validators[:100] {
		abcivals[i] = val.ABCIValidatorUpdate(suite.stakingKeeper.PowerReduction(suite.ctx))
	}

	// remove genesis validator
	vals = vals[:100]
	suite.Require().Equal(abcivals, vals)
}
