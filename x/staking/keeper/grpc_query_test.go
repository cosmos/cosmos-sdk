package keeper_test

import (
	gocontext "context"
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type QueryTestSuite struct {
	suite.Suite

	app         *simapp.SimApp
	ctx         sdk.Context
	addrs       []sdk.AccAddress
	vals        []types.Validator
	queryClient types.QueryClient
}

func (suite *QueryTestSuite) SetupTest() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, abci.Header{})

	querier := keeper.Querier{Keeper: app.StakingKeeper}

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, querier)
	queryClient := types.NewQueryClient(queryHelper)

	addrs, _ := createValidators(ctx, app, []int64{9, 8, 7})
	validators := app.StakingKeeper.GetValidators(ctx, 5)

	header := abci.Header{
		ChainID: "HelloChain",
		Height:  5,
	}

	hi := types.NewHistoricalInfo(header, validators)
	app.StakingKeeper.SetHistoricalInfo(ctx, 5, hi)

	suite.app, suite.ctx, suite.queryClient, suite.addrs, suite.vals = app, ctx, queryClient, addrs, validators
}

func (suite *QueryTestSuite) TestGRPCQueryValidators() {
	queryClient, vals := suite.queryClient, suite.vals

	valsResp, err := queryClient.Validators(gocontext.Background(), &types.QueryValidatorsRequest{})
	suite.Error(err)
	suite.Nil(vals)

	req := &types.QueryValidatorsRequest{Status: sdk.Bonded.String(),
		Req: &query.PageRequest{Limit: 1, CountTotal: true}}
	valsResp, err = queryClient.Validators(gocontext.Background(), req)

	suite.NoError(err)
	suite.Equal(1, len(valsResp.Validators))
	suite.NotNil(valsResp.Res.NextKey)
	suite.Equal(uint64(len(vals)), valsResp.Res.Total)

	valRes, err := queryClient.Validator(gocontext.Background(), &types.QueryValidatorRequest{})
	suite.Error(err)
	suite.Nil(valRes)

	valRes, err = queryClient.Validator(gocontext.Background(), &types.QueryValidatorRequest{ValidatorAddr: valsResp.Validators[0].OperatorAddress})
	suite.NoError(err)
	suite.Equal(valsResp.Validators[0], valRes.Validator)
}

func (suite *QueryTestSuite) TestDelegatorValidators() {
	app, ctx, queryClient, addrs := suite.app, suite.ctx, suite.queryClient, suite.addrs
	params := app.StakingKeeper.GetParams(ctx)

	res, err := queryClient.DelegatorValidators(gocontext.Background(), &types.QueryDelegatorValidatorsRequest{})
	suite.Error(err)
	suite.Nil(res)

	delValidators := app.StakingKeeper.GetDelegatorValidators(ctx, addrs[0], params.MaxValidators)

	delegatorParamsReq := &types.QueryDelegatorValidatorsRequest{
		DelegatorAddr: addrs[0],
		Req:           &query.PageRequest{Limit: 1, CountTotal: true},
	}

	res, err = queryClient.DelegatorValidators(gocontext.Background(), delegatorParamsReq)
	suite.NoError(err)
	suite.Equal(1, len(res.Validators))
	suite.NotNil(res.Res.NextKey)
	suite.Equal(uint64(len(delValidators)), res.Res.Total)
}

func (suite *QueryTestSuite) TestGRPCQueryDelegatorValidator() {
	queryClient, addrs, vals := suite.queryClient, suite.addrs, suite.vals

	addrVal1 := vals[1].OperatorAddress
	valResp, err := queryClient.DelegatorValidator(gocontext.Background(), &types.QueryDelegatorValidatorRequest{})
	suite.Error(err)
	suite.Nil(valResp)

	req := &types.QueryDelegatorValidatorRequest{DelegatorAddr: addrs[1], ValidatorAddr: addrVal1}
	valResp, err = queryClient.DelegatorValidator(gocontext.Background(), req)
	suite.NoError(err)
	suite.Equal(addrVal1, valResp.Validator.OperatorAddress)
}

func (suite *QueryTestSuite) TestGRPCQueryDelegation() {
	app, ctx, queryClient, addrs, vals := suite.app, suite.ctx, suite.queryClient, suite.addrs, suite.vals
	addrAcc := addrs[0]
	addrVal := vals[0].OperatorAddress

	delegationResp, err := queryClient.Delegation(gocontext.Background(), &types.QueryDelegationRequest{})
	suite.Error(err)
	suite.Nil(delegationResp)

	delReq := &types.QueryDelegationRequest{DelegatorAddr: addrAcc, ValidatorAddr: addrVal}
	delegation, found := app.StakingKeeper.GetDelegation(ctx, addrAcc, addrVal)
	suite.True(found)

	delegationResp, err = queryClient.Delegation(gocontext.Background(), delReq)
	suite.NoError(err)

	suite.Equal(delegation.ValidatorAddress, delegationResp.DelegationResponse.Delegation.ValidatorAddress)
	suite.Equal(delegation.DelegatorAddress, delegationResp.DelegationResponse.Delegation.DelegatorAddress)
	suite.Equal(sdk.NewCoin(sdk.DefaultBondDenom, delegation.Shares.TruncateInt()), delegationResp.DelegationResponse.Balance)
}

func (suite *QueryTestSuite) TestGRPCDelegatorDelegations() {
	app, ctx, queryClient, addrs, vals := suite.app, suite.ctx, suite.queryClient, suite.addrs, suite.vals
	addrAcc := addrs[0]
	addrVal1 := vals[0].OperatorAddress

	delegation, found := app.StakingKeeper.GetDelegation(ctx, addrAcc, addrVal1)
	suite.True(found)

	delegatorDel, err := queryClient.DelegatorDelegations(gocontext.Background(), &types.QueryDelegatorDelegationsRequest{})
	suite.Error(err)
	suite.Nil(delegatorDel)

	// Query delegator delegations with pagination
	delDelReq := &types.QueryDelegatorDelegationsRequest{DelegatorAddr: addrAcc,
		Req: &query.PageRequest{Limit: 1, CountTotal: true}}
	delegatorDelegationsResp, err := queryClient.DelegatorDelegations(gocontext.Background(), delDelReq)
	suite.NoError(err)

	suite.Equal(uint64(2), delegatorDelegationsResp.Res.Total)
	suite.Len(delegatorDelegationsResp.DelegationResponses, 1)
	suite.Equal(1, len(delegatorDelegationsResp.DelegationResponses))
	suite.Equal(sdk.NewCoin(sdk.DefaultBondDenom, delegation.Shares.TruncateInt()), delegatorDelegationsResp.DelegationResponses[0].Balance)
}

func (suite *QueryTestSuite) TestGRPCValidatorDelegations() {
	app, ctx, queryClient, addrs, vals := suite.app, suite.ctx, suite.queryClient, suite.addrs, suite.vals
	addrAcc := addrs[0]
	addrVal1 := vals[0].OperatorAddress

	validatorDelegations, err := queryClient.ValidatorDelegations(gocontext.Background(), &types.QueryValidatorDelegationsRequest{})
	suite.Error(err)
	suite.Nil(validatorDelegations)

	delegation, found := app.StakingKeeper.GetDelegation(ctx, addrAcc, addrVal1)
	suite.True(found)
	valReq := &types.QueryValidatorDelegationsRequest{ValidatorAddr: addrVal1,
		Req: &query.PageRequest{Limit: 1, CountTotal: true}}
	delegationsRes, err := queryClient.ValidatorDelegations(gocontext.Background(), valReq)

	suite.NoError(err)
	suite.Len(delegationsRes.DelegationResponses, 1)
	suite.NotNil(delegationsRes.Res.NextKey)
	// suite.Equal(uint64(2), delegationsRes.Res.Total)
	suite.Equal(addrVal1, delegationsRes.DelegationResponses[0].Delegation.ValidatorAddress)
	suite.Equal(sdk.NewCoin(sdk.DefaultBondDenom, delegation.Shares.TruncateInt()), delegationsRes.DelegationResponses[0].Balance)
}

func (suite *QueryTestSuite) TestGRPCQueryUnbondingDelegation() {
	app, ctx, queryClient, addrs, vals := suite.app, suite.ctx, suite.queryClient, suite.addrs, suite.vals
	addrAcc2 := addrs[1]
	addrVal2 := vals[1].OperatorAddress

	unbondingTokens := sdk.TokensFromConsensusPower(2)
	_, err := app.StakingKeeper.Undelegate(ctx, addrAcc2, addrVal2, unbondingTokens.ToDec())
	suite.NoError(err)

	unbondRes, err := queryClient.UnbondingDelegation(gocontext.Background(), &types.QueryUnbondingDelegationRequest{})
	suite.Error(err)
	suite.Nil(unbondRes)

	unbondReq := &types.QueryUnbondingDelegationRequest{DelegatorAddr: addrAcc2, ValidatorAddr: addrVal2}
	unbondRes, err = queryClient.UnbondingDelegation(gocontext.Background(), unbondReq)

	unbond, found := app.StakingKeeper.GetUnbondingDelegation(ctx, addrAcc2, addrVal2)
	suite.True(found)
	suite.NotNil(unbondRes)
	suite.Equal(unbond, unbondRes.Unbond)
}

func (suite *QueryTestSuite) TestGRPCQueryDelegatorUnbondingDelegations() {
	app, ctx, queryClient, addrs, vals := suite.app, suite.ctx, suite.queryClient, suite.addrs, suite.vals
	addrAcc := addrs[0]
	addrVal, addrVal2 := vals[0].OperatorAddress, vals[1].OperatorAddress

	unbondingTokens := sdk.TokensFromConsensusPower(2)
	_, err := app.StakingKeeper.Undelegate(ctx, addrAcc, addrVal, unbondingTokens.ToDec())
	suite.NoError(err)
	_, err = app.StakingKeeper.Undelegate(ctx, addrAcc, addrVal2, unbondingTokens.ToDec())
	suite.NoError(err)

	delegatorUbdsRes, err := queryClient.DelegatorUnbondingDelegations(gocontext.Background(),
		&types.QueryDelegatorUnbondingDelegationsRequest{})
	suite.Error(err)
	suite.Nil(delegatorUbdsRes)

	unbond, found := app.StakingKeeper.GetUnbondingDelegation(ctx, addrAcc, addrVal)
	suite.True(found)
	// Query Delegator unbonding delegate with pagination
	unbReq := &types.QueryDelegatorUnbondingDelegationsRequest{DelegatorAddr: addrAcc,
		Req: &query.PageRequest{Limit: 1, CountTotal: true}}
	delegatorUbdsRes, err = queryClient.DelegatorUnbondingDelegations(gocontext.Background(), unbReq)

	suite.NoError(err)
	suite.NotNil(delegatorUbdsRes.Res.NextKey)
	suite.Equal(uint64(2), delegatorUbdsRes.Res.Total)
	suite.Len(delegatorUbdsRes.UnbondingResponses, 1)
	suite.Equal(unbond, delegatorUbdsRes.UnbondingResponses[0])
}

func (suite *QueryTestSuite) TestGRPCQueryPoolParameters() {
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
	resp, err := queryClient.Parameters(gocontext.Background(), &types.QueryParametersRequest{})
	suite.NoError(err)
	suite.Equal(app.StakingKeeper.GetParams(ctx), resp.Params)
}

func (suite *QueryTestSuite) TestGRPCHistoricalInfo() {
	app, ctx, queryClient := suite.app, suite.ctx, suite.queryClient

	hi, found := app.StakingKeeper.GetHistoricalInfo(ctx, 5)
	suite.True(found)
	hist, err := queryClient.HistoricalInfo(gocontext.Background(), &types.QueryHistoricalInfoRequest{})
	suite.Error(err)
	suite.Nil(hist)

	hist, err = queryClient.HistoricalInfo(gocontext.Background(), &types.QueryHistoricalInfoRequest{Height: 4})
	suite.Error(err)
	suite.Nil(hist)

	hist, err = queryClient.HistoricalInfo(gocontext.Background(), &types.QueryHistoricalInfoRequest{Height: 5})
	suite.NoError(err)
	suite.NotNil(hist)
	suite.Equal(&hi, hist.Hist)
}

func (suite *QueryTestSuite) TestGRPCQueryRedelegation() {
	app, ctx, queryClient, addrs, vals := suite.app, suite.ctx, suite.queryClient, suite.addrs, suite.vals

	_, addrAcc := addrs[0], addrs[1]
	val1, val2 := vals[0], vals[1]
	delAmount := sdk.TokensFromConsensusPower(1)
	_, err := app.StakingKeeper.Delegate(ctx, addrAcc, delAmount, sdk.Unbonded, val1, true)
	suite.NoError(err)
	_ = app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)

	rdAmount := sdk.TokensFromConsensusPower(1)
	_, err = app.StakingKeeper.BeginRedelegation(ctx, addrAcc, val1.GetOperator(), val2.GetOperator(), rdAmount.ToDec())
	suite.NoError(err)
	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)

	redel, found := app.StakingKeeper.GetRedelegation(ctx, addrAcc, val1.OperatorAddress, val2.OperatorAddress)
	suite.True(found)

	redelReq := &types.QueryRedelegationsRequest{
		DelegatorAddr: addrAcc, SrcValidatorAddr: val1.OperatorAddress, DstValidatorAddr: val2.OperatorAddress,
		Req: &query.PageRequest{}}
	redelResp, err := queryClient.Redelegations(gocontext.Background(), redelReq)

	suite.NoError(err)
	suite.Len(redelResp.RedelegationResponses, 1)
	suite.Equal(redel.DelegatorAddress, redelResp.RedelegationResponses[0].Redelegation.DelegatorAddress)
	suite.Equal(redel.ValidatorSrcAddress, redelResp.RedelegationResponses[0].Redelegation.ValidatorSrcAddress)
	suite.Equal(redel.ValidatorDstAddress, redelResp.RedelegationResponses[0].Redelegation.ValidatorDstAddress)
	suite.Len(redel.Entries, len(redelResp.RedelegationResponses[0].Entries))

	// delegator redelegations
	redelResp, err = queryClient.Redelegations(gocontext.Background(), &types.QueryRedelegationsRequest{
		DelegatorAddr: addrAcc, SrcValidatorAddr: val1.OperatorAddress, Req: &query.PageRequest{}})
	suite.NoError(err)

	suite.Len(redelResp.RedelegationResponses, len(redel.Entries))
	suite.Equal(redel.DelegatorAddress, redelResp.RedelegationResponses[0].Redelegation.DelegatorAddress)
	suite.Equal(redel.ValidatorSrcAddress, redelResp.RedelegationResponses[0].Redelegation.ValidatorSrcAddress)
	suite.Equal(redel.ValidatorDstAddress, redelResp.RedelegationResponses[0].Redelegation.ValidatorDstAddress)
	suite.Len(redel.Entries, len(redelResp.RedelegationResponses[0].Entries))

	redelResp, err = queryClient.Redelegations(gocontext.Background(), &types.QueryRedelegationsRequest{
		SrcValidatorAddr: val1.GetOperator(), Req: &query.PageRequest{Limit: 1, CountTotal: true}})

	suite.NoError(err)
	suite.Equal(uint64(1), redelResp.Res.Total)
	suite.Len(redelResp.RedelegationResponses, 1)
	suite.Equal(redel.DelegatorAddress, redelResp.RedelegationResponses[0].Redelegation.DelegatorAddress)
	suite.Equal(redel.ValidatorSrcAddress, redelResp.RedelegationResponses[0].Redelegation.ValidatorSrcAddress)
	suite.Equal(redel.ValidatorDstAddress, redelResp.RedelegationResponses[0].Redelegation.ValidatorDstAddress)
	suite.Len(redel.Entries, len(redelResp.RedelegationResponses[0].Entries))
}

func (suite *QueryTestSuite) TestUnbondingDelegation() {
	app, ctx, queryClient, addrs, vals := suite.app, suite.ctx, suite.queryClient, suite.addrs, suite.vals

	addrAcc1, addrAcc2 := addrs[0], addrs[1]
	val1 := vals[0]

	// undelegate
	undelAmount := sdk.TokensFromConsensusPower(2)
	_, err := app.StakingKeeper.Undelegate(ctx, addrAcc1, val1.GetOperator(), undelAmount.ToDec())
	suite.NoError(err)
	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)

	unbDelsResp, err := queryClient.UnbondingDelegation(gocontext.Background(), &types.QueryUnbondingDelegationRequest{
		DelegatorAddr: addrAcc2, ValidatorAddr: val1.GetOperator()})
	suite.Error(err)

	unbDelsResp, err = queryClient.UnbondingDelegation(gocontext.Background(), &types.QueryUnbondingDelegationRequest{
		DelegatorAddr: addrAcc1, ValidatorAddr: val1.GetOperator()})

	suite.Equal(addrAcc1, unbDelsResp.Unbond.DelegatorAddress)
	suite.Equal(val1.OperatorAddress, unbDelsResp.Unbond.ValidatorAddress)
	suite.Equal(1, len(unbDelsResp.Unbond.Entries))
}

func (suite *QueryTestSuite) TestGRPCQueryValidatorUnbondingDelegations() {
	app, ctx, queryClient, addrs, vals := suite.app, suite.ctx, suite.queryClient, suite.addrs, suite.vals
	addrAcc1, _ := addrs[0], addrs[1]
	val1 := vals[0]

	// undelegate
	undelAmount := sdk.TokensFromConsensusPower(2)
	_, err := app.StakingKeeper.Undelegate(ctx, addrAcc1, val1.GetOperator(), undelAmount.ToDec())
	suite.NoError(err)
	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)

	valUnbonds, err := queryClient.ValidatorUnbondingDelegations(gocontext.Background(), &types.QueryValidatorUnbondingDelegationsRequest{})
	suite.Error(err)
	suite.Nil(valUnbonds)

	valUnbonds, err = queryClient.ValidatorUnbondingDelegations(gocontext.Background(),
		&types.QueryValidatorUnbondingDelegationsRequest{ValidatorAddr: val1.GetOperator(),
			Req: &query.PageRequest{Limit: 1, CountTotal: true}})
	suite.NoError(err)
	suite.Equal(uint64(1), valUnbonds.Res.Total)
	suite.Equal(1, len(valUnbonds.UnbondingResponses))
}

func createValidators(ctx sdk.Context, app *simapp.SimApp, powers []int64) ([]sdk.AccAddress, []sdk.ValAddress) {
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

	return addrs, valAddrs
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(QueryTestSuite))
}
