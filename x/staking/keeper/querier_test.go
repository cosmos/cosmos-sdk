package keeper_test

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (suite *KeeperTestSuite) TestNewQuerier() {
	addrs := simtestutil.AddTestAddrs(suite.bankKeeper, suite.stakingKeeper, suite.ctx, 500, sdk.NewInt(10000))
	_, addrAcc2 := addrs[0], addrs[1]
	addrVal1, _ := sdk.ValAddress(addrs[0]), sdk.ValAddress(addrs[1])

	// Create Validators
	amts := []sdk.Int{sdk.NewInt(9), sdk.NewInt(8)}
	var validators [2]types.Validator
	for i, amt := range amts {
		validators[i] = teststaking.NewValidator(suite.T(), sdk.ValAddress(addrs[i]), PKs[i])
		validators[i], _ = validators[i].AddTokensFromDel(amt)
		suite.stakingKeeper.SetValidator(suite.ctx, validators[i])
		suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, validators[i])
	}

	header := tmproto.Header{
		ChainID: "HelloChain",
		Height:  5,
	}
	hi := types.NewHistoricalInfo(header, validators[:], suite.stakingKeeper.PowerReduction(suite.ctx))
	suite.stakingKeeper.SetHistoricalInfo(suite.ctx, 5, &hi)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	legacyQuerierCdc := codec.NewAminoCodec(suite.legacyAmino)
	querier := keeper.NewQuerier(suite.stakingKeeper, legacyQuerierCdc.LegacyAmino)

	bz, err := querier(suite.ctx, []string{"other"}, query)
	suite.Require().Error(err)
	suite.Require().Nil(bz)

	_, err = querier(suite.ctx, []string{"pool"}, query)
	suite.Require().NoError(err)

	_, err = querier(suite.ctx, []string{"parameters"}, query)
	suite.Require().NoError(err)

	queryValParams := types.NewQueryValidatorParams(addrVal1, 0, 0)
	bz, errRes := suite.legacyAmino.MarshalJSON(queryValParams)
	suite.Require().NoError(errRes)

	query.Path = "/custom/staking/validator"
	query.Data = bz

	_, err = querier(suite.ctx, []string{"validator"}, query)
	suite.Require().NoError(err)

	_, err = querier(suite.ctx, []string{"validatorDelegations"}, query)
	suite.Require().NoError(err)

	_, err = querier(suite.ctx, []string{"validatorUnbondingDelegations"}, query)
	suite.Require().NoError(err)

	queryDelParams := types.NewQueryDelegatorParams(addrAcc2)
	bz, errRes = suite.legacyAmino.MarshalJSON(queryDelParams)
	suite.Require().NoError(errRes)

	query.Path = "/custom/staking/validator"
	query.Data = bz

	_, err = querier(suite.ctx, []string{"delegatorDelegations"}, query)
	suite.Require().NoError(err)

	_, err = querier(suite.ctx, []string{"delegatorUnbondingDelegations"}, query)
	suite.Require().NoError(err)

	_, err = querier(suite.ctx, []string{"delegatorValidators"}, query)
	suite.Require().NoError(err)

	bz, errRes = suite.legacyAmino.MarshalJSON(types.NewQueryRedelegationParams(nil, nil, nil))
	suite.Require().NoError(errRes)
	query.Data = bz

	_, err = querier(suite.ctx, []string{"redelegations"}, query)
	suite.Require().NoError(err)

	queryHisParams := types.QueryHistoricalInfoRequest{Height: 5}
	bz, errRes = suite.legacyAmino.MarshalJSON(queryHisParams)
	suite.Require().NoError(errRes)

	query.Path = "/custom/staking/historicalInfo"
	query.Data = bz

	_, err = querier(suite.ctx, []string{"historicalInfo"}, query)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestQueryParametersPool() {
	legacyQuerierCdc := codec.NewAminoCodec(suite.legacyAmino)
	querier := keeper.NewQuerier(suite.stakingKeeper, legacyQuerierCdc.LegacyAmino)

	bondDenom := sdk.DefaultBondDenom

	res, err := querier(suite.ctx, []string{types.QueryParameters}, abci.RequestQuery{})
	suite.Require().NoError(err)

	var params types.Params
	errRes := suite.legacyAmino.UnmarshalJSON(res, &params)
	suite.Require().NoError(errRes)
	suite.Require().Equal(suite.stakingKeeper.GetParams(suite.ctx), params)

	res, err = querier(suite.ctx, []string{types.QueryPool}, abci.RequestQuery{})
	suite.Require().NoError(err)

	var pool types.Pool
	bondedPool := suite.stakingKeeper.GetBondedPool(suite.ctx)
	notBondedPool := suite.stakingKeeper.GetNotBondedPool(suite.ctx)
	suite.Require().NoError(suite.legacyAmino.UnmarshalJSON(res, &pool))
	suite.Require().Equal(suite.bankKeeper.GetBalance(suite.ctx, notBondedPool.GetAddress(), bondDenom).Amount, pool.NotBondedTokens)
	suite.Require().Equal(suite.bankKeeper.GetBalance(suite.ctx, bondedPool.GetAddress(), bondDenom).Amount, pool.BondedTokens)
}

func (suite *KeeperTestSuite) TestQueryValidators() {
	params := suite.stakingKeeper.GetParams(suite.ctx)
	legacyQuerierCdc := codec.NewAminoCodec(suite.legacyAmino)
	querier := keeper.NewQuerier(suite.stakingKeeper, legacyQuerierCdc.LegacyAmino)

	addrs := simtestutil.AddTestAddrs(suite.bankKeeper, suite.stakingKeeper, suite.ctx, 500, suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 10000))

	// Create Validators
	amts := []sdk.Int{sdk.NewInt(8), sdk.NewInt(7)}
	status := []types.BondStatus{types.Unbonded, types.Unbonding}
	var validators [2]types.Validator
	for i, amt := range amts {
		validators[i] = teststaking.NewValidator(suite.T(), sdk.ValAddress(addrs[i]), PKs[i])
		validators[i], _ = validators[i].AddTokensFromDel(amt)
		validators[i] = validators[i].UpdateStatus(status[i])
	}

	suite.stakingKeeper.SetValidator(suite.ctx, validators[0])
	suite.stakingKeeper.SetValidator(suite.ctx, validators[1])

	// Query Validators
	queriedValidators := suite.stakingKeeper.GetValidators(suite.ctx, params.MaxValidators)
	suite.Require().Len(queriedValidators, 3)

	for i, s := range status {
		queryValsParams := types.NewQueryValidatorsParams(1, int(params.MaxValidators), s.String())
		bz, err := suite.legacyAmino.MarshalJSON(queryValsParams)
		suite.Require().NoError(err)

		req := abci.RequestQuery{
			Path: fmt.Sprintf("/custom/%s/%s", types.QuerierRoute, types.QueryValidators),
			Data: bz,
		}

		res, err := querier(suite.ctx, []string{types.QueryValidators}, req)
		suite.Require().NoError(err)

		var validatorsResp []types.Validator
		err = suite.legacyAmino.UnmarshalJSON(res, &validatorsResp)
		suite.Require().NoError(err)

		suite.Require().Equal(1, len(validatorsResp))
		suite.Require().Equal(validators[i].OperatorAddress, validatorsResp[0].OperatorAddress)
	}

	// Query each validator
	for _, validator := range validators {
		queryParams := types.NewQueryValidatorParams(validator.GetOperator(), 0, 0)
		bz, err := suite.legacyAmino.MarshalJSON(queryParams)
		suite.Require().NoError(err)

		query := abci.RequestQuery{
			Path: "/custom/staking/validator",
			Data: bz,
		}
		res, err := querier(suite.ctx, []string{types.QueryValidator}, query)
		suite.Require().NoError(err)

		var queriedValidator types.Validator
		err = suite.legacyAmino.UnmarshalJSON(res, &queriedValidator)
		suite.Require().NoError(err)

		suite.Require().True(validator.Equal(&queriedValidator))
	}
}

func (suite *KeeperTestSuite) TestQueryDelegation() {
	params := suite.stakingKeeper.GetParams(suite.ctx)
	legacyQuerierCdc := codec.NewAminoCodec(suite.legacyAmino)
	querier := keeper.NewQuerier(suite.stakingKeeper, legacyQuerierCdc.LegacyAmino)

	addrs := simtestutil.AddTestAddrs(suite.bankKeeper, suite.stakingKeeper, suite.ctx, 2, suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 10000))
	addrAcc1, addrAcc2 := addrs[0], addrs[1]
	addrVal1, addrVal2 := sdk.ValAddress(addrAcc1), sdk.ValAddress(addrAcc2)

	pubKeys := simtestutil.CreateTestPubKeys(2)
	pk1, pk2 := pubKeys[0], pubKeys[1]

	// Create Validators and Delegation
	val1 := teststaking.NewValidator(suite.T(), addrVal1, pk1)
	suite.stakingKeeper.SetValidator(suite.ctx, val1)
	suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, val1)

	val2 := teststaking.NewValidator(suite.T(), addrVal2, pk2)
	suite.stakingKeeper.SetValidator(suite.ctx, val2)
	suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, val2)

	delTokens := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 20)
	_, err := suite.stakingKeeper.Delegate(suite.ctx, addrAcc2, delTokens, types.Unbonded, val1, true)
	suite.Require().NoError(err)

	// apply TM updates
	applyValidatorSetUpdates(suite.T(), suite.ctx, suite.stakingKeeper, -1)

	// Query Delegator bonded validators
	queryParams := types.NewQueryDelegatorParams(addrAcc2)
	bz, errRes := suite.legacyAmino.MarshalJSON(queryParams)
	suite.Require().NoError(errRes)

	query := abci.RequestQuery{
		Path: "/custom/staking/delegatorValidators",
		Data: bz,
	}

	delValidators := suite.stakingKeeper.GetDelegatorValidators(suite.ctx, addrAcc2, params.MaxValidators)

	res, err := querier(suite.ctx, []string{types.QueryDelegatorValidators}, query)
	suite.Require().NoError(err)

	var validatorsResp types.Validators
	errRes = suite.legacyAmino.UnmarshalJSON(res, &validatorsResp)
	suite.Require().NoError(errRes)

	suite.Require().Equal(len(delValidators), len(validatorsResp))
	suite.Require().ElementsMatch(delValidators, validatorsResp)

	// error unknown request
	query.Data = bz[:len(bz)-1]

	_, err = querier(suite.ctx, []string{types.QueryDelegatorValidators}, query)
	suite.Require().Error(err)

	// Query bonded validator
	queryBondParams := types.QueryDelegatorValidatorRequest{DelegatorAddr: addrAcc2.String(), ValidatorAddr: addrVal1.String()}
	bz, errRes = suite.legacyAmino.MarshalJSON(queryBondParams)
	suite.Require().NoError(errRes)

	query = abci.RequestQuery{
		Path: "/custom/staking/delegatorValidator",
		Data: bz,
	}

	res, err = querier(suite.ctx, []string{types.QueryDelegatorValidator}, query)
	suite.Require().NoError(err)

	var validator types.Validator
	errRes = suite.legacyAmino.UnmarshalJSON(res, &validator)
	suite.Require().NoError(errRes)
	suite.Require().True(validator.Equal(&delValidators[0]))

	// error unknown request
	query.Data = bz[:len(bz)-1]

	_, err = querier(suite.ctx, []string{types.QueryDelegatorValidator}, query)
	suite.Require().Error(err)

	// Query delegation

	query = abci.RequestQuery{
		Path: "/custom/staking/delegation",
		Data: bz,
	}

	delegation, found := suite.stakingKeeper.GetDelegation(suite.ctx, addrAcc2, addrVal1)
	suite.Require().True(found)

	res, err = querier(suite.ctx, []string{types.QueryDelegation}, query)
	suite.Require().NoError(err)

	var delegationRes types.DelegationResponse
	errRes = suite.legacyAmino.UnmarshalJSON(res, &delegationRes)
	suite.Require().NoError(errRes)

	suite.Require().Equal(delegation.ValidatorAddress, delegationRes.Delegation.ValidatorAddress)
	suite.Require().Equal(delegation.DelegatorAddress, delegationRes.Delegation.DelegatorAddress)
	suite.Require().Equal(sdk.NewCoin(sdk.DefaultBondDenom, delegation.Shares.TruncateInt()), delegationRes.Balance)

	// Query Delegator Delegations
	bz, errRes = suite.legacyAmino.MarshalJSON(queryParams)
	suite.Require().NoError(errRes)

	query = abci.RequestQuery{
		Path: "/custom/staking/delegatorDelegations",
		Data: bz,
	}

	res, err = querier(suite.ctx, []string{types.QueryDelegatorDelegations}, query)
	suite.Require().NoError(err)

	var delegatorDelegations types.DelegationResponses
	errRes = suite.legacyAmino.UnmarshalJSON(res, &delegatorDelegations)
	suite.Require().NoError(errRes)
	suite.Require().Len(delegatorDelegations, 1)
	suite.Require().Equal(delegation.ValidatorAddress, delegatorDelegations[0].Delegation.ValidatorAddress)
	suite.Require().Equal(delegation.DelegatorAddress, delegatorDelegations[0].Delegation.DelegatorAddress)
	suite.Require().Equal(sdk.NewCoin(sdk.DefaultBondDenom, delegation.Shares.TruncateInt()), delegatorDelegations[0].Balance)

	// error unknown request
	query.Data = bz[:len(bz)-1]

	_, err = querier(suite.ctx, []string{types.QueryDelegation}, query)
	suite.Require().Error(err)

	// Query validator delegations
	bz, errRes = suite.legacyAmino.MarshalJSON(types.NewQueryValidatorParams(addrVal1, 1, 100))
	suite.Require().NoError(errRes)

	query = abci.RequestQuery{
		Path: "custom/staking/validatorDelegations",
		Data: bz,
	}

	res, err = querier(suite.ctx, []string{types.QueryValidatorDelegations}, query)
	suite.Require().NoError(err)

	var delegationsRes types.DelegationResponses
	errRes = suite.legacyAmino.UnmarshalJSON(res, &delegationsRes)
	suite.Require().NoError(errRes)
	suite.Require().Len(delegatorDelegations, 1)
	suite.Require().Equal(delegation.ValidatorAddress, delegationsRes[0].Delegation.ValidatorAddress)
	suite.Require().Equal(delegation.DelegatorAddress, delegationsRes[0].Delegation.DelegatorAddress)
	suite.Require().Equal(sdk.NewCoin(sdk.DefaultBondDenom, delegation.Shares.TruncateInt()), delegationsRes[0].Balance)

	// Query unbonding delegation
	unbondingTokens := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 10)
	_, err = suite.stakingKeeper.Undelegate(suite.ctx, addrAcc2, val1.GetOperator(), sdk.NewDecFromInt(unbondingTokens))
	suite.Require().NoError(err)

	queryBondParams = types.QueryDelegatorValidatorRequest{DelegatorAddr: addrAcc2.String(), ValidatorAddr: addrVal1.String()}
	bz, errRes = suite.legacyAmino.MarshalJSON(queryBondParams)
	suite.Require().NoError(errRes)

	query = abci.RequestQuery{
		Path: "/custom/staking/unbondingDelegation",
		Data: bz,
	}

	unbond, found := suite.stakingKeeper.GetUnbondingDelegation(suite.ctx, addrAcc2, addrVal1)
	suite.Require().True(found)

	res, err = querier(suite.ctx, []string{types.QueryUnbondingDelegation}, query)
	suite.Require().NoError(err)

	var unbondRes types.UnbondingDelegation
	errRes = suite.legacyAmino.UnmarshalJSON(res, &unbondRes)
	suite.Require().NoError(errRes)

	suite.Require().Equal(unbond, unbondRes)

	// error unknown request
	query.Data = bz[:len(bz)-1]

	_, err = querier(suite.ctx, []string{types.QueryUnbondingDelegation}, query)
	suite.Require().Error(err)

	// Query Delegator Unbonding Delegations

	query = abci.RequestQuery{
		Path: "/custom/staking/delegatorUnbondingDelegations",
		Data: bz,
	}

	res, err = querier(suite.ctx, []string{types.QueryDelegatorUnbondingDelegations}, query)
	suite.Require().NoError(err)

	var delegatorUbds []types.UnbondingDelegation
	errRes = suite.legacyAmino.UnmarshalJSON(res, &delegatorUbds)
	suite.Require().NoError(errRes)
	suite.Require().Equal(unbond, delegatorUbds[0])

	// error unknown request
	query.Data = bz[:len(bz)-1]

	_, err = querier(suite.ctx, []string{types.QueryDelegatorUnbondingDelegations}, query)
	suite.Require().Error(err)

	// Query redelegation
	redelegationTokens := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 10)
	_, err = suite.stakingKeeper.BeginRedelegation(suite.ctx, addrAcc2, val1.GetOperator(), val2.GetOperator(), sdk.NewDecFromInt(redelegationTokens))
	suite.Require().NoError(err)
	redel, found := suite.stakingKeeper.GetRedelegation(suite.ctx, addrAcc2, val1.GetOperator(), val2.GetOperator())
	suite.Require().True(found)

	bz, errRes = suite.legacyAmino.MarshalJSON(types.NewQueryRedelegationParams(addrAcc2, val1.GetOperator(), val2.GetOperator()))
	suite.Require().NoError(errRes)

	query = abci.RequestQuery{
		Path: "/custom/staking/redelegations",
		Data: bz,
	}

	res, err = querier(suite.ctx, []string{types.QueryRedelegations}, query)
	suite.Require().NoError(err)

	var redelRes types.RedelegationResponses
	errRes = suite.legacyAmino.UnmarshalJSON(res, &redelRes)
	suite.Require().NoError(errRes)
	suite.Require().Len(redelRes, 1)
	suite.Require().Equal(redel.DelegatorAddress, redelRes[0].Redelegation.DelegatorAddress)
	suite.Require().Equal(redel.ValidatorSrcAddress, redelRes[0].Redelegation.ValidatorSrcAddress)
	suite.Require().Equal(redel.ValidatorDstAddress, redelRes[0].Redelegation.ValidatorDstAddress)
	suite.Require().Len(redel.Entries, len(redelRes[0].Entries))
}

func (suite *KeeperTestSuite) TestQueryValidatorDelegations_Pagination() {
	cases := []struct {
		page            int
		limit           int
		expectedResults int
	}{
		{
			page:            1,
			limit:           75,
			expectedResults: 75,
		},
		{
			page:            2,
			limit:           75,
			expectedResults: 25,
		},
		{
			page:            1,
			limit:           100,
			expectedResults: 100,
		},
	}

	legacyQuerierCdc := codec.NewAminoCodec(suite.legacyAmino)
	querier := keeper.NewQuerier(suite.stakingKeeper, legacyQuerierCdc.LegacyAmino)

	addrs := simtestutil.AddTestAddrs(suite.bankKeeper, suite.stakingKeeper, suite.ctx, 100, suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 10000))
	pubKeys := simtestutil.CreateTestPubKeys(1)

	valAddress := sdk.ValAddress(addrs[0])

	val1 := teststaking.NewValidator(suite.T(), valAddress, pubKeys[0])
	suite.stakingKeeper.SetValidator(suite.ctx, val1)
	suite.stakingKeeper.SetValidatorByPowerIndex(suite.ctx, val1)

	// Create Validators and Delegation
	for _, addr := range addrs {
		validator, found := suite.stakingKeeper.GetValidator(suite.ctx, valAddress)
		if !found {
			suite.T().Error("expected validator not found")
		}

		delTokens := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 20)
		_, err := suite.stakingKeeper.Delegate(suite.ctx, addr, delTokens, types.Unbonded, validator, true)
		suite.Require().NoError(err)
	}

	// apply TM updates
	applyValidatorSetUpdates(suite.T(), suite.ctx, suite.stakingKeeper, -1)

	for _, c := range cases {
		// Query Delegator bonded validators
		queryParams := types.NewQueryDelegatorParams(addrs[0])
		bz, errRes := suite.legacyAmino.MarshalJSON(queryParams)
		suite.Require().NoError(errRes)

		// Query valAddress delegations
		bz, errRes = suite.legacyAmino.MarshalJSON(types.NewQueryValidatorParams(valAddress, c.page, c.limit))
		suite.Require().NoError(errRes)

		query := abci.RequestQuery{
			Path: "custom/staking/validatorDelegations",
			Data: bz,
		}

		res, err := querier(suite.ctx, []string{types.QueryValidatorDelegations}, query)
		suite.Require().NoError(err)

		var delegationsRes types.DelegationResponses
		errRes = suite.legacyAmino.UnmarshalJSON(res, &delegationsRes)
		suite.Require().NoError(errRes)
		suite.Require().Len(delegationsRes, c.expectedResults)
	}

	// Undelegate
	for _, addr := range addrs {
		delTokens := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 20)
		_, err := suite.stakingKeeper.Undelegate(suite.ctx, addr, val1.GetOperator(), sdk.NewDecFromInt(delTokens))
		suite.Require().NoError(err)
	}

	// apply TM updates
	applyValidatorSetUpdates(suite.T(), suite.ctx, suite.stakingKeeper, -1)

	for _, c := range cases {
		// Query Unbonding delegations with pagination.
		queryParams := types.NewQueryDelegatorParams(addrs[0])
		bz, errRes := suite.legacyAmino.MarshalJSON(queryParams)
		suite.Require().NoError(errRes)

		bz, errRes = suite.legacyAmino.MarshalJSON(types.NewQueryValidatorParams(valAddress, c.page, c.limit))
		suite.Require().NoError(errRes)
		query := abci.RequestQuery{
			Data: bz,
		}

		unbondingDelegations := types.UnbondingDelegations{}
		res, err := querier(suite.ctx, []string{types.QueryValidatorUnbondingDelegations}, query)
		suite.Require().NoError(err)

		errRes = suite.legacyAmino.UnmarshalJSON(res, &unbondingDelegations)
		suite.Require().NoError(errRes)
		suite.Require().Len(unbondingDelegations, c.expectedResults)
	}
}

func (suite *KeeperTestSuite) TestQueryRedelegations() {
	legacyQuerierCdc := codec.NewAminoCodec(suite.legacyAmino)
	querier := keeper.NewQuerier(suite.stakingKeeper, legacyQuerierCdc.LegacyAmino)

	addrs := simtestutil.AddTestAddrs(suite.bankKeeper, suite.stakingKeeper, suite.ctx, 2, suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 10000))
	addrAcc1, addrAcc2 := addrs[0], addrs[1]
	addrVal1, addrVal2 := sdk.ValAddress(addrAcc1), sdk.ValAddress(addrAcc2)

	// Create Validators and Delegation
	val1 := teststaking.NewValidator(suite.T(), addrVal1, PKs[0])
	val2 := teststaking.NewValidator(suite.T(), addrVal2, PKs[1])
	suite.stakingKeeper.SetValidator(suite.ctx, val1)
	suite.stakingKeeper.SetValidator(suite.ctx, val2)

	delAmount := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 100)
	_, err := suite.stakingKeeper.Delegate(suite.ctx, addrAcc2, delAmount, types.Unbonded, val1, true)
	suite.Require().NoError(err)
	applyValidatorSetUpdates(suite.T(), suite.ctx, suite.stakingKeeper, -1)

	rdAmount := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 20)
	_, err = suite.stakingKeeper.BeginRedelegation(suite.ctx, addrAcc2, val1.GetOperator(), val2.GetOperator(), sdk.NewDecFromInt(rdAmount))
	suite.Require().NoError(err)
	applyValidatorSetUpdates(suite.T(), suite.ctx, suite.stakingKeeper, -1)

	redel, found := suite.stakingKeeper.GetRedelegation(suite.ctx, addrAcc2, val1.GetOperator(), val2.GetOperator())
	suite.Require().True(found)

	// delegator redelegations
	queryDelegatorParams := types.NewQueryDelegatorParams(addrAcc2)
	bz, errRes := suite.legacyAmino.MarshalJSON(queryDelegatorParams)
	suite.Require().NoError(errRes)

	query := abci.RequestQuery{
		Path: "/custom/staking/redelegations",
		Data: bz,
	}

	res, err := querier(suite.ctx, []string{types.QueryRedelegations}, query)
	suite.Require().NoError(err)

	var redelRes types.RedelegationResponses
	errRes = suite.legacyAmino.UnmarshalJSON(res, &redelRes)
	suite.Require().NoError(errRes)
	suite.Require().Len(redelRes, 1)
	suite.Require().Equal(redel.DelegatorAddress, redelRes[0].Redelegation.DelegatorAddress)
	suite.Require().Equal(redel.ValidatorSrcAddress, redelRes[0].Redelegation.ValidatorSrcAddress)
	suite.Require().Equal(redel.ValidatorDstAddress, redelRes[0].Redelegation.ValidatorDstAddress)
	suite.Require().Len(redel.Entries, len(redelRes[0].Entries))

	// validator redelegations
	queryValidatorParams := types.NewQueryValidatorParams(val1.GetOperator(), 0, 0)
	bz, errRes = suite.legacyAmino.MarshalJSON(queryValidatorParams)
	suite.Require().NoError(errRes)

	query = abci.RequestQuery{
		Path: "/custom/staking/redelegations",
		Data: bz,
	}

	res, err = querier(suite.ctx, []string{types.QueryRedelegations}, query)
	suite.Require().NoError(err)

	errRes = suite.legacyAmino.UnmarshalJSON(res, &redelRes)
	suite.Require().NoError(errRes)
	suite.Require().Len(redelRes, 1)
	suite.Require().Equal(redel.DelegatorAddress, redelRes[0].Redelegation.DelegatorAddress)
	suite.Require().Equal(redel.ValidatorSrcAddress, redelRes[0].Redelegation.ValidatorSrcAddress)
	suite.Require().Equal(redel.ValidatorDstAddress, redelRes[0].Redelegation.ValidatorDstAddress)
	suite.Require().Len(redel.Entries, len(redelRes[0].Entries))
}

func (suite *KeeperTestSuite) TestQueryUnbondingDelegation() {
	legacyQuerierCdc := codec.NewAminoCodec(suite.legacyAmino)
	querier := keeper.NewQuerier(suite.stakingKeeper, legacyQuerierCdc.LegacyAmino)

	addrs := simtestutil.AddTestAddrs(suite.bankKeeper, suite.stakingKeeper, suite.ctx, 2, suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 10000))
	addrAcc1, addrAcc2 := addrs[0], addrs[1]
	addrVal1 := sdk.ValAddress(addrAcc1)

	// Create Validators and Delegation
	val1 := teststaking.NewValidator(suite.T(), addrVal1, PKs[0])
	suite.stakingKeeper.SetValidator(suite.ctx, val1)

	// delegate
	delAmount := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 100)
	_, err := suite.stakingKeeper.Delegate(suite.ctx, addrAcc1, delAmount, types.Unbonded, val1, true)
	suite.Require().NoError(err)
	applyValidatorSetUpdates(suite.T(), suite.ctx, suite.stakingKeeper, -1)

	// undelegate
	undelAmount := suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 20)
	_, err = suite.stakingKeeper.Undelegate(suite.ctx, addrAcc1, val1.GetOperator(), sdk.NewDecFromInt(undelAmount))
	suite.Require().NoError(err)
	applyValidatorSetUpdates(suite.T(), suite.ctx, suite.stakingKeeper, -1)

	_, found := suite.stakingKeeper.GetUnbondingDelegation(suite.ctx, addrAcc1, val1.GetOperator())
	suite.Require().True(found)

	//
	// found: query unbonding delegation by delegator and validator
	//
	queryValidatorParams := types.QueryDelegatorValidatorRequest{DelegatorAddr: addrAcc1.String(), ValidatorAddr: val1.GetOperator().String()}
	bz, errRes := suite.legacyAmino.MarshalJSON(queryValidatorParams)
	suite.Require().NoError(errRes)
	query := abci.RequestQuery{
		Path: "/custom/staking/unbondingDelegation",
		Data: bz,
	}
	res, err := querier(suite.ctx, []string{types.QueryUnbondingDelegation}, query)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	var ubDel types.UnbondingDelegation
	suite.Require().NoError(suite.legacyAmino.UnmarshalJSON(res, &ubDel))
	suite.Require().Equal(addrAcc1.String(), ubDel.DelegatorAddress)
	suite.Require().Equal(val1.OperatorAddress, ubDel.ValidatorAddress)
	suite.Require().Equal(1, len(ubDel.Entries))

	//
	// not found: query unbonding delegation by delegator and validator
	//
	queryValidatorParams = types.QueryDelegatorValidatorRequest{DelegatorAddr: addrAcc2.String(), ValidatorAddr: val1.GetOperator().String()}
	bz, errRes = suite.legacyAmino.MarshalJSON(queryValidatorParams)
	suite.Require().NoError(errRes)
	query = abci.RequestQuery{
		Path: "/custom/staking/unbondingDelegation",
		Data: bz,
	}
	_, err = querier(suite.ctx, []string{types.QueryUnbondingDelegation}, query)
	suite.Require().Error(err)

	//
	// found: query unbonding delegation by delegator and validator
	//
	queryDelegatorParams := types.NewQueryDelegatorParams(addrAcc1)
	bz, errRes = suite.legacyAmino.MarshalJSON(queryDelegatorParams)
	suite.Require().NoError(errRes)
	query = abci.RequestQuery{
		Path: "/custom/staking/delegatorUnbondingDelegations",
		Data: bz,
	}
	res, err = querier(suite.ctx, []string{types.QueryDelegatorUnbondingDelegations}, query)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	var ubDels []types.UnbondingDelegation
	suite.Require().NoError(suite.legacyAmino.UnmarshalJSON(res, &ubDels))
	suite.Require().Equal(1, len(ubDels))
	suite.Require().Equal(addrAcc1.String(), ubDels[0].DelegatorAddress)
	suite.Require().Equal(val1.OperatorAddress, ubDels[0].ValidatorAddress)

	//
	// not found: query unbonding delegation by delegator and validator
	//
	queryDelegatorParams = types.NewQueryDelegatorParams(addrAcc2)
	bz, errRes = suite.legacyAmino.MarshalJSON(queryDelegatorParams)
	suite.Require().NoError(errRes)
	query = abci.RequestQuery{
		Path: "/custom/staking/delegatorUnbondingDelegations",
		Data: bz,
	}
	res, err = querier(suite.ctx, []string{types.QueryDelegatorUnbondingDelegations}, query)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().NoError(suite.legacyAmino.UnmarshalJSON(res, &ubDels))
	suite.Require().Equal(0, len(ubDels))
}

func (suite *KeeperTestSuite) TestQueryHistoricalInfo() {
	legacyQuerierCdc := codec.NewAminoCodec(suite.legacyAmino)
	querier := keeper.NewQuerier(suite.stakingKeeper, legacyQuerierCdc.LegacyAmino)

	addrs := simtestutil.AddTestAddrs(suite.bankKeeper, suite.stakingKeeper, suite.ctx, 2, suite.stakingKeeper.TokensFromConsensusPower(suite.ctx, 10000))
	addrAcc1, addrAcc2 := addrs[0], addrs[1]
	addrVal1, addrVal2 := sdk.ValAddress(addrAcc1), sdk.ValAddress(addrAcc2)

	// Create Validators and Delegation
	val1 := teststaking.NewValidator(suite.T(), addrVal1, PKs[0])
	val2 := teststaking.NewValidator(suite.T(), addrVal2, PKs[1])
	vals := []types.Validator{val1, val2}
	suite.stakingKeeper.SetValidator(suite.ctx, val1)
	suite.stakingKeeper.SetValidator(suite.ctx, val2)

	header := tmproto.Header{
		ChainID: "HelloChain",
		Height:  5,
	}
	hi := types.NewHistoricalInfo(header, vals, suite.stakingKeeper.PowerReduction(suite.ctx))
	suite.stakingKeeper.SetHistoricalInfo(suite.ctx, 5, &hi)

	queryHistoricalParams := types.QueryHistoricalInfoRequest{Height: 4}
	bz, errRes := suite.legacyAmino.MarshalJSON(queryHistoricalParams)
	suite.Require().NoError(errRes)
	query := abci.RequestQuery{
		Path: "/custom/staking/historicalInfo",
		Data: bz,
	}
	res, err := querier(suite.ctx, []string{types.QueryHistoricalInfo}, query)
	suite.Require().Error(err, "Invalid query passed")
	suite.Require().Nil(res, "Invalid query returned non-nil result")

	queryHistoricalParams = types.QueryHistoricalInfoRequest{Height: 5}
	bz, errRes = suite.legacyAmino.MarshalJSON(queryHistoricalParams)
	suite.Require().NoError(errRes)
	query.Data = bz
	res, err = querier(suite.ctx, []string{types.QueryHistoricalInfo}, query)
	suite.Require().NoError(err, "Valid query passed")
	suite.Require().NotNil(res, "Valid query returned nil result")

	var recv types.HistoricalInfo
	suite.Require().NoError(suite.legacyAmino.UnmarshalJSON(res, &recv))
	suite.Require().Equal(hi, recv, "HistoricalInfo query returned wrong result")
}
