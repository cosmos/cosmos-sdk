package keeper_test

import (
	gocontext "context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (suite *KeeperTestSuite) TestGRPCQueryValidators() {
	queryClient, vals := suite.queryClient, suite.vals
	var req *types.QueryValidatorsRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = &types.QueryValidatorsRequest{}
			},
			false,
		},
		{"invalid request with empty status",
			func() {
				req = &types.QueryValidatorsRequest{Status: ""}
			},
			false,
		},
		{
			"invalid request",
			func() {
				req = &types.QueryValidatorsRequest{Status: "test"}
			},
			false,
		},
		{"valid request",
			func() {
				req = &types.QueryValidatorsRequest{Status: sdk.Bonded.String(),
					Req: &query.PageRequest{Limit: 1, CountTotal: true}}
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			valsResp, err := queryClient.Validators(gocontext.Background(), req)
			if tc.expPass {
				suite.NoError(err)
				suite.NotNil(valsResp)
				suite.Equal(1, len(valsResp.Validators))
				suite.NotNil(valsResp.Res.NextKey)
				suite.Equal(uint64(len(vals)), valsResp.Res.Total)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCValidator() {
	app, ctx, queryClient, vals := suite.app, suite.ctx, suite.queryClient, suite.vals
	validator, found := app.StakingKeeper.GetValidator(ctx, vals[0].OperatorAddress)
	suite.True(found)
	var req *types.QueryValidatorRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = &types.QueryValidatorRequest{}
			},
			false,
		},
		{"valid request",
			func() {
				req = &types.QueryValidatorRequest{ValidatorAddr: vals[0].OperatorAddress}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			res, err := queryClient.Validator(gocontext.Background(), req)
			if tc.expPass {
				suite.NoError(err)
				suite.Equal(validator, res.Validator)
			} else {
				suite.Error(err)
				suite.Nil(res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCQueryDelegatorValidators() {
	app, ctx, queryClient, addrs := suite.app, suite.ctx, suite.queryClient, suite.addrs
	params := app.StakingKeeper.GetParams(ctx)
	delValidators := app.StakingKeeper.GetDelegatorValidators(ctx, addrs[0], params.MaxValidators)
	var req *types.QueryDelegatorValidatorsRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = &types.QueryDelegatorValidatorsRequest{}
			},
			false,
		},
		{"valid request",
			func() {
				req = &types.QueryDelegatorValidatorsRequest{
					DelegatorAddr: addrs[0],
					Req:           &query.PageRequest{Limit: 1, CountTotal: true}}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			res, err := queryClient.DelegatorValidators(gocontext.Background(), req)
			if tc.expPass {
				suite.NoError(err)
				suite.Equal(1, len(res.Validators))
				suite.NotNil(res.Res.NextKey)
				suite.Equal(uint64(len(delValidators)), res.Res.Total)
			} else {
				suite.Error(err)
				suite.Nil(res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCQueryDelegatorValidator() {
	queryClient, addrs, vals := suite.queryClient, suite.addrs, suite.vals
	addr := addrs[1]
	addrVal, addrVal1 := vals[0].OperatorAddress, vals[1].OperatorAddress
	var req *types.QueryDelegatorValidatorRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = &types.QueryDelegatorValidatorRequest{}
			},
			false,
		},
		{"invalid delegator, validator pair",
			func() {
				req = &types.QueryDelegatorValidatorRequest{
					DelegatorAddr: addr,
					ValidatorAddr: addrVal,
				}
			},
			false,
		},
		{"valid request",
			func() {
				req = &types.QueryDelegatorValidatorRequest{
					DelegatorAddr: addr,
					ValidatorAddr: addrVal1,
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			res, err := queryClient.DelegatorValidator(gocontext.Background(), req)
			if tc.expPass {
				suite.NoError(err)
				suite.Equal(addrVal1, res.Validator.OperatorAddress)
			} else {
				suite.Error(err)
				suite.Nil(res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCQueryDelegation() {
	app, ctx, queryClient, addrs, vals := suite.app, suite.ctx, suite.queryClient, suite.addrs, suite.vals
	addrAcc, addrAcc1 := addrs[0], addrs[1]
	addrVal := vals[0].OperatorAddress

	delegation, found := app.StakingKeeper.GetDelegation(ctx, addrAcc, addrVal)
	suite.True(found)
	var req *types.QueryDelegationRequest

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"empty request",
			func() {
				req = &types.QueryDelegationRequest{}
			},
			false,
		},
		{"invalid validator, delegator pair",
			func() {
				req = &types.QueryDelegationRequest{
					DelegatorAddr: addrAcc1,
					ValidatorAddr: addrVal,
				}
			},
			false,
		},
		{"valid request",
			func() {
				req = &types.QueryDelegationRequest{DelegatorAddr: addrAcc, ValidatorAddr: addrVal}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			res, err := queryClient.Delegation(gocontext.Background(), req)
			if tc.expPass {
				suite.Equal(delegation.ValidatorAddress, res.DelegationResponse.Delegation.ValidatorAddress)
				suite.Equal(delegation.DelegatorAddress, res.DelegationResponse.Delegation.DelegatorAddress)
				suite.Equal(sdk.NewCoin(sdk.DefaultBondDenom, delegation.Shares.TruncateInt()), res.DelegationResponse.Balance)
			} else {
				suite.Error(err)
				suite.Nil(res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCQueryDelegatorDelegations() {
	app, ctx, queryClient, addrs, vals := suite.app, suite.ctx, suite.queryClient, suite.addrs, suite.vals
	addrAcc := addrs[0]
	addrVal1 := vals[0].OperatorAddress

	delegation, found := app.StakingKeeper.GetDelegation(ctx, addrAcc, addrVal1)
	suite.True(found)
	var req *types.QueryDelegatorDelegationsRequest

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"empty request",
			func() {
				req = &types.QueryDelegatorDelegationsRequest{}
			},
			false,
		}, {"invalid request",
			func() {
				req = &types.QueryDelegatorDelegationsRequest{DelegatorAddr: addrs[4]}
			},
			false,
		},
		{"valid request",
			func() {
				req = &types.QueryDelegatorDelegationsRequest{DelegatorAddr: addrAcc,
					Req: &query.PageRequest{Limit: 1, CountTotal: true}}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			res, err := queryClient.DelegatorDelegations(gocontext.Background(), req)
			if tc.expPass {
				suite.Equal(uint64(2), res.Res.Total)
				suite.Len(res.DelegationResponses, 1)
				suite.Equal(1, len(res.DelegationResponses))
				suite.Equal(sdk.NewCoin(sdk.DefaultBondDenom, delegation.Shares.TruncateInt()), res.DelegationResponses[0].Balance)
			} else {
				suite.Error(err)
				suite.Nil(res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCQueryValidatorDelegations() {
	app, ctx, queryClient, addrs, vals := suite.app, suite.ctx, suite.queryClient, suite.addrs, suite.vals
	addrAcc := addrs[0]
	addrVal1 := vals[1].OperatorAddress
	valAddrs := simapp.ConvertAddrsToValAddrs(addrs)
	addrVal2 := valAddrs[4]

	delegation, found := app.StakingKeeper.GetDelegation(ctx, addrAcc, addrVal1)
	suite.True(found)

	var req *types.QueryValidatorDelegationsRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		expErr   bool
	}{
		{"empty request",
			func() {
				req = &types.QueryValidatorDelegationsRequest{}
			},
			false,
			true,
		},
		{"invalid validator delegator pair",
			func() {
				req = &types.QueryValidatorDelegationsRequest{ValidatorAddr: addrVal2}
			},
			false,
			false,
		},
		{"valid request",
			func() {
				req = &types.QueryValidatorDelegationsRequest{ValidatorAddr: addrVal1,
					Req: &query.PageRequest{Limit: 1, CountTotal: true}}
			},
			true,
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			res, err := queryClient.ValidatorDelegations(gocontext.Background(), req)
			if tc.expPass && !tc.expErr {
				suite.NoError(err)
				suite.Len(res.DelegationResponses, 1)
				suite.NotNil(res.Res.NextKey)
				suite.Equal(uint64(2), res.Res.Total)
				suite.Equal(addrVal1, res.DelegationResponses[0].Delegation.ValidatorAddress)
				suite.Equal(sdk.NewCoin(sdk.DefaultBondDenom, delegation.Shares.TruncateInt()), res.DelegationResponses[0].Balance)
			} else if !tc.expPass && !tc.expErr {
				suite.NoError(err)
				suite.Nil(res.DelegationResponses)
			} else {
				suite.Error(err)
				suite.Nil(res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCQueryUnbondingDelegation() {
	app, ctx, queryClient, addrs, vals := suite.app, suite.ctx, suite.queryClient, suite.addrs, suite.vals
	addrAcc2 := addrs[1]
	addrVal2 := vals[1].OperatorAddress

	unbondingTokens := sdk.TokensFromConsensusPower(2)
	_, err := app.StakingKeeper.Undelegate(ctx, addrAcc2, addrVal2, unbondingTokens.ToDec())
	suite.NoError(err)

	unbond, found := app.StakingKeeper.GetUnbondingDelegation(ctx, addrAcc2, addrVal2)
	suite.True(found)
	var req *types.QueryUnbondingDelegationRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"empty request",
			func() {
				req = &types.QueryUnbondingDelegationRequest{}
			},
			false,
		},
		{"invalid request",
			func() {
				req = &types.QueryUnbondingDelegationRequest{}
			},
			false,
		},
		{"valid request",
			func() {
				req = &types.QueryUnbondingDelegationRequest{
					DelegatorAddr: addrAcc2, ValidatorAddr: addrVal2}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			res, err := queryClient.UnbondingDelegation(gocontext.Background(), req)
			if tc.expPass {
				suite.NotNil(res)
				suite.Equal(unbond, res.Unbond)
			} else {
				suite.Error(err)
				suite.Nil(res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCQueryDelegatorUnbondingDelegations() {
	app, ctx, queryClient, addrs, vals := suite.app, suite.ctx, suite.queryClient, suite.addrs, suite.vals
	addrAcc, addrAcc1 := addrs[0], addrs[1]
	addrVal, addrVal2 := vals[0].OperatorAddress, vals[1].OperatorAddress

	unbondingTokens := sdk.TokensFromConsensusPower(2)
	_, err := app.StakingKeeper.Undelegate(ctx, addrAcc, addrVal, unbondingTokens.ToDec())
	suite.NoError(err)
	_, err = app.StakingKeeper.Undelegate(ctx, addrAcc, addrVal2, unbondingTokens.ToDec())
	suite.NoError(err)

	unbond, found := app.StakingKeeper.GetUnbondingDelegation(ctx, addrAcc, addrVal)
	suite.True(found)
	var req *types.QueryDelegatorUnbondingDelegationsRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		expErr   bool
	}{
		{"empty request",
			func() {
				req = &types.QueryDelegatorUnbondingDelegationsRequest{}
			},
			false,
			true,
		},
		{"invalid request",
			func() {
				req = &types.QueryDelegatorUnbondingDelegationsRequest{DelegatorAddr: addrAcc1}
			},
			false,
			false,
		},
		{"valid request",
			func() {
				req = &types.QueryDelegatorUnbondingDelegationsRequest{DelegatorAddr: addrAcc,
					Req: &query.PageRequest{Limit: 1, CountTotal: true}}
			},
			true,
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			res, err := queryClient.DelegatorUnbondingDelegations(gocontext.Background(), req)
			if tc.expPass && !tc.expErr {
				suite.NoError(err)
				suite.NotNil(res.Res.NextKey)
				suite.Equal(uint64(2), res.Res.Total)
				suite.Len(res.UnbondingResponses, 1)
				suite.Equal(unbond, res.UnbondingResponses[0])
			} else if !tc.expPass && !tc.expErr {
				suite.NoError(err)
				suite.Nil(res.UnbondingResponses)
			} else {
				suite.Error(err)
				suite.Nil(res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCQueryPoolParameters() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient
	bondDenom := sdk.DefaultBondDenom

	// Query pool
	res, err := queryClient.Pool(gocontext.Background(), &types.QueryPoolRequest{})
	suite.NoError(err)
	bondedPool := app.StakingKeeper.GetBondedPool(ctx)
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
	suite.Equal(app.BankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), bondDenom).Amount, res.Pool.NotBondedTokens)
	suite.Equal(app.BankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount, res.Pool.BondedTokens)

	// Query Params
	resp, err := queryClient.Params(gocontext.Background(), &types.QueryParamsRequest{})
	suite.NoError(err)
	suite.Equal(app.StakingKeeper.GetParams(ctx), resp.Params)
}

func (suite *KeeperTestSuite) TestGRPCQueryHistoricalInfo() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient

	hi, found := app.StakingKeeper.GetHistoricalInfo(ctx, 5)
	suite.True(found)

	var req *types.QueryHistoricalInfoRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"empty request",
			func() {
				req = &types.QueryHistoricalInfoRequest{}
			},
			false,
		},
		{"invalid request with negative height",
			func() {
				req = &types.QueryHistoricalInfoRequest{Height: -1}
			},
			false,
		},
		{"valid request with old height",
			func() {
				req = &types.QueryHistoricalInfoRequest{Height: 4}
			},
			false,
		},
		{"valid request with current height",
			func() {
				req = &types.QueryHistoricalInfoRequest{Height: 5}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			res, err := queryClient.HistoricalInfo(gocontext.Background(), req)
			if tc.expPass {
				suite.NoError(err)
				suite.NotNil(res)
				suite.Equal(&hi, res.Hist)
			} else {
				suite.Error(err)
				suite.Nil(res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCQueryRedelegation() {
	app, ctx, queryClient, addrs, vals := suite.app, suite.ctx, suite.queryClient, suite.addrs, suite.vals

	addrAcc, addrAcc1 := addrs[0], addrs[1]
	valAddrs := simapp.ConvertAddrsToValAddrs(addrs)
	val1, val2, val3, val4 := vals[0], vals[1], valAddrs[3], valAddrs[4]
	delAmount := sdk.TokensFromConsensusPower(1)
	_, err := app.StakingKeeper.Delegate(ctx, addrAcc1, delAmount, sdk.Unbonded, val1, true)
	suite.NoError(err)
	_ = app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)

	rdAmount := sdk.TokensFromConsensusPower(1)
	_, err = app.StakingKeeper.BeginRedelegation(ctx, addrAcc1, val1.GetOperator(), val2.GetOperator(), rdAmount.ToDec())
	suite.NoError(err)
	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)

	redel, found := app.StakingKeeper.GetRedelegation(ctx, addrAcc1, val1.OperatorAddress, val2.OperatorAddress)
	suite.True(found)

	var req *types.QueryRedelegationsRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
		expErr   bool
	}{
		{"request redelegations for non existant addr",
			func() {
				req = &types.QueryRedelegationsRequest{DelegatorAddr: addrAcc}
			},
			false,
			false,
		},
		{"request redelegations with non existent pairs",
			func() {
				req = &types.QueryRedelegationsRequest{DelegatorAddr: addrAcc, SrcValidatorAddr: val3,
					DstValidatorAddr: val4}
			},
			false,
			true,
		},
		{"request redelegations with delegatoraddr, sourceValAddr, destValAddr",
			func() {
				req = &types.QueryRedelegationsRequest{
					DelegatorAddr: addrAcc1, SrcValidatorAddr: val1.OperatorAddress,
					DstValidatorAddr: val2.OperatorAddress, Req: &query.PageRequest{}}
			},
			true,
			false,
		},
		{"request redelegations with delegatoraddr and sourceValAddr",
			func() {
				req = &types.QueryRedelegationsRequest{
					DelegatorAddr: addrAcc1, SrcValidatorAddr: val1.OperatorAddress,
					Req: &query.PageRequest{}}
			},
			true,
			false,
		},
		{"query redelegations with sourceValAddr only",
			func() {
				req = &types.QueryRedelegationsRequest{
					SrcValidatorAddr: val1.GetOperator(),
					Req:              &query.PageRequest{Limit: 1, CountTotal: true}}
			},
			true,
			false,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			res, err := queryClient.Redelegations(gocontext.Background(), req)
			if tc.expPass && !tc.expErr {
				suite.NoError(err)
				suite.Len(res.RedelegationResponses, len(redel.Entries))
				suite.Equal(redel.DelegatorAddress, res.RedelegationResponses[0].Redelegation.DelegatorAddress)
				suite.Equal(redel.ValidatorSrcAddress, res.RedelegationResponses[0].Redelegation.ValidatorSrcAddress)
				suite.Equal(redel.ValidatorDstAddress, res.RedelegationResponses[0].Redelegation.ValidatorDstAddress)
				suite.Len(redel.Entries, len(res.RedelegationResponses[0].Entries))
			} else if !tc.expPass && !tc.expErr {
				suite.NoError(err)
				suite.Nil(res.RedelegationResponses)
			} else {
				suite.Error(err)
				suite.Nil(res)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGRPCQueryValidatorUnbondingDelegations() {
	app, ctx, queryClient, addrs, vals := suite.app, suite.ctx, suite.queryClient, suite.addrs, suite.vals
	addrAcc1, _ := addrs[0], addrs[1]
	val1 := vals[0]

	// undelegate
	undelAmount := sdk.TokensFromConsensusPower(2)
	_, err := app.StakingKeeper.Undelegate(ctx, addrAcc1, val1.GetOperator(), undelAmount.ToDec())
	suite.NoError(err)
	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)

	var req *types.QueryValidatorUnbondingDelegationsRequest
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{"empty request",
			func() {
				req = &types.QueryValidatorUnbondingDelegationsRequest{}
			},
			false,
		},
		{"valid request",
			func() {
				req = &types.QueryValidatorUnbondingDelegationsRequest{
					ValidatorAddr: val1.GetOperator(),
					Req:           &query.PageRequest{Limit: 1, CountTotal: true}}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			res, err := queryClient.ValidatorUnbondingDelegations(gocontext.Background(), req)
			if tc.expPass {
				suite.NoError(err)
				suite.Equal(uint64(1), res.Res.Total)
				suite.Equal(1, len(res.UnbondingResponses))
			} else {
				suite.Error(err)
				suite.Nil(res)
			}
		})
	}
}

func createValidators(ctx sdk.Context, app *simapp.SimApp, powers []int64) ([]sdk.AccAddress, []sdk.ValAddress, []types.Validator) {
	addrs := simapp.AddTestAddrsIncremental(app, ctx, 5, sdk.NewInt(300000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addrs)
	pks := simapp.CreateTestPubKeys(5)

	appCodec, _ := simapp.MakeCodecs()
	app.StakingKeeper = keeper.NewKeeper(
		appCodec,
		app.GetKey(types.StoreKey),
		app.AccountKeeper,
		app.BankKeeper,
		app.GetSubspace(types.ModuleName),
	)

	val1 := types.NewValidator(valAddrs[0], pks[0], types.Description{})
	val2 := types.NewValidator(valAddrs[1], pks[1], types.Description{})
	vals := []types.Validator{val1, val2}

	app.StakingKeeper.SetValidator(ctx, val1)
	app.StakingKeeper.SetValidator(ctx, val2)
	app.StakingKeeper.SetValidatorByConsAddr(ctx, val1)
	app.StakingKeeper.SetValidatorByConsAddr(ctx, val2)
	app.StakingKeeper.SetNewValidatorByPowerIndex(ctx, val1)
	app.StakingKeeper.SetNewValidatorByPowerIndex(ctx, val2)

	_, _ = app.StakingKeeper.Delegate(ctx, addrs[0], sdk.TokensFromConsensusPower(powers[0]), sdk.Unbonded, val1, true)
	_, _ = app.StakingKeeper.Delegate(ctx, addrs[1], sdk.TokensFromConsensusPower(powers[1]), sdk.Unbonded, val2, true)
	_, _ = app.StakingKeeper.Delegate(ctx, addrs[0], sdk.TokensFromConsensusPower(powers[2]), sdk.Unbonded, val2, true)
	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)

	return addrs, valAddrs, vals
}
