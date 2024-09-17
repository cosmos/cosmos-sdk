package keeper_test

import (
	gocontext "context"
	"fmt"
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	"cosmossdk.io/x/staking/types"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
)

func createValidatorAccs(t *testing.T, f *fixture) ([]sdk.AccAddress, []types.Validator) {
	t.Helper()
	addrs, _, validators := createValidators(&testing.T{}, f, []int64{9, 8, 7})

	// sort a copy of the validators, so that original validators does not
	// have its order changed
	sortedVals := make([]types.Validator, len(validators))
	copy(sortedVals, validators)

	return addrs, validators
}

func TestGRPCQueryValidators(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	_, vals := createValidatorAccs(t, f)

	qr := f.app.QueryHelper()
	queryClient := types.NewQueryClient(qr)

	var req *types.QueryValidatorsRequest
	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		numVals   int
		hasNext   bool
		expErrMsg string
	}{
		{
			"empty request",
			func() {
				req = &types.QueryValidatorsRequest{}
			},
			true,
			len(vals),
			false,
			"",
		},
		{
			"empty status returns all the validators",
			func() {
				req = &types.QueryValidatorsRequest{Status: ""}
			},
			true,
			len(vals),
			false,
			"",
		},
		{
			"invalid request",
			func() {
				req = &types.QueryValidatorsRequest{Status: "test"}
			},
			false,
			0,
			false,
			"invalid validator status test",
		},
		{
			"valid request",
			func() {
				req = &types.QueryValidatorsRequest{
					Status:     types.Bonded.String(),
					Pagination: &query.PageRequest{Limit: 1, CountTotal: true},
				}
			},
			true,
			1,
			true,
			"",
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.msg), func(t *testing.T) {
			tc.malleate()
			valsResp, err := queryClient.Validators(gocontext.Background(), req)
			if tc.expPass {
				assert.NilError(t, err)
				assert.Assert(t, valsResp != nil)
				assert.Equal(t, tc.numVals, len(valsResp.Validators))
				assert.Equal(t, uint64(len(vals)), valsResp.Pagination.Total)

				if tc.hasNext {
					assert.Assert(t, valsResp.Pagination.NextKey != nil)
				} else {
					assert.Assert(t, valsResp.Pagination.NextKey == nil)
				}
			} else {
				assert.ErrorContains(t, err, tc.expErrMsg)
			}
		})
	}
}

func TestGRPCQueryDelegatorValidators(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx
	addrs, _ := createValidatorAccs(t, f)

	qr := f.app.QueryHelper()
	queryClient := types.NewQueryClient(qr)

	params, err := f.stakingKeeper.Params.Get(ctx)
	assert.NilError(t, err)
	delValidators, err := f.stakingKeeper.GetDelegatorValidators(ctx, addrs[0], params.MaxValidators)
	assert.NilError(t, err)
	var req *types.QueryDelegatorValidatorsRequest
	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		expErrMsg string
	}{
		{
			"empty request",
			func() {
				req = &types.QueryDelegatorValidatorsRequest{}
			},
			false,
			"delegator address cannot be empty",
		},
		{
			"invalid delegator address",
			func() {
				req = &types.QueryDelegatorValidatorsRequest{
					DelegatorAddr: "invalid",
					Pagination:    &query.PageRequest{Limit: 1, CountTotal: true},
				}
			},
			false,
			"invalid bech32",
		},
		{
			"valid request",
			func() {
				req = &types.QueryDelegatorValidatorsRequest{
					DelegatorAddr: addrs[0].String(),
					Pagination:    &query.PageRequest{Limit: 1, CountTotal: true},
				}
			},
			true,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.msg), func(t *testing.T) {
			tc.malleate()
			res, err := queryClient.DelegatorValidators(gocontext.Background(), req)
			if tc.expPass {
				assert.NilError(t, err)
				assert.Equal(t, 1, len(res.Validators))
				assert.Assert(t, res.Pagination.NextKey != nil)
				assert.Equal(t, uint64(len(delValidators.Validators)), res.Pagination.Total)
			} else {
				assert.ErrorContains(t, err, tc.expErrMsg)
				assert.Assert(t, res == nil)
			}
		})
	}
}

func TestGRPCQueryDelegatorValidator(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	addrs, vals := createValidatorAccs(t, f)

	qr := f.app.QueryHelper()
	queryClient := types.NewQueryClient(qr)

	addr := addrs[1]
	addrVal, addrVal1 := vals[0].OperatorAddress, vals[1].OperatorAddress
	var req *types.QueryDelegatorValidatorRequest
	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		expErrMsg string
	}{
		{
			"empty request",
			func() {
				req = &types.QueryDelegatorValidatorRequest{}
			},
			false,
			"delegator address cannot be empty",
		},
		{
			"invalid delegator, validator pair",
			func() {
				req = &types.QueryDelegatorValidatorRequest{
					DelegatorAddr: addr.String(),
					ValidatorAddr: addrVal,
				}
			},
			false,
			"not found",
		},
		{
			"empty delegator address",
			func() {
				req = &types.QueryDelegatorValidatorRequest{
					DelegatorAddr: "",
					ValidatorAddr: addrVal1,
				}
			},
			false,
			"delegator address cannot be empty",
		},
		{
			"empty validator address",
			func() {
				req = &types.QueryDelegatorValidatorRequest{
					DelegatorAddr: addr.String(),
					ValidatorAddr: "",
				}
			},
			false,
			"validator address cannot be empty",
		},
		{
			"valid request",
			func() {
				req = &types.QueryDelegatorValidatorRequest{
					DelegatorAddr: addr.String(),
					ValidatorAddr: addrVal1,
				}
			},
			true,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.msg), func(t *testing.T) {
			tc.malleate()
			res, err := queryClient.DelegatorValidator(gocontext.Background(), req)
			if tc.expPass {
				assert.NilError(t, err)
				assert.Equal(t, addrVal1, res.Validator.OperatorAddress)
			} else {
				assert.ErrorContains(t, err, tc.expErrMsg)
				assert.Assert(t, res == nil)
			}
		})
	}
}

func TestGRPCQueryDelegation(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx
	addrs, vals := createValidatorAccs(t, f)

	qr := f.app.QueryHelper()
	queryClient := types.NewQueryClient(qr)

	addrAcc, addrAcc1 := addrs[0], addrs[1]
	addrVal := vals[0].OperatorAddress
	valAddr, err := sdk.ValAddressFromBech32(addrVal)
	assert.NilError(t, err)
	delegation, found := f.stakingKeeper.Delegations.Get(ctx, collections.Join(addrAcc, valAddr))
	assert.Assert(t, found)
	var req *types.QueryDelegationRequest

	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		expErrMsg string
	}{
		{
			"empty request",
			func() {
				req = &types.QueryDelegationRequest{}
			},
			false,
			"delegator address cannot be empty",
		},
		{
			"invalid validator, delegator pair",
			func() {
				req = &types.QueryDelegationRequest{
					DelegatorAddr: addrAcc1.String(),
					ValidatorAddr: addrVal,
				}
			},
			false,
			fmt.Sprintf("delegation with delegator %s not found for validator %s", addrAcc1.String(), addrVal),
		},
		{
			"valid request",
			func() {
				req = &types.QueryDelegationRequest{DelegatorAddr: addrAcc.String(), ValidatorAddr: addrVal}
			},
			true,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.msg), func(t *testing.T) {
			tc.malleate()
			res, err := queryClient.Delegation(gocontext.Background(), req)
			if tc.expPass {
				assert.Equal(t, delegation.ValidatorAddress, res.DelegationResponse.Delegation.ValidatorAddress)
				assert.Equal(t, delegation.DelegatorAddress, res.DelegationResponse.Delegation.DelegatorAddress)
				assert.DeepEqual(t, sdk.NewCoin(sdk.DefaultBondDenom, delegation.Shares.TruncateInt()), res.DelegationResponse.Balance)
			} else {
				assert.ErrorContains(t, err, tc.expErrMsg)
				assert.Assert(t, res == nil)
			}
		})
	}
}

func TestGRPCQueryDelegatorDelegations(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx
	addrs, vals := createValidatorAccs(t, f)

	qr := f.app.QueryHelper()
	queryClient := types.NewQueryClient(qr)

	addrAcc := addrs[0]
	addrVal1 := vals[0].OperatorAddress
	valAddr, err := sdk.ValAddressFromBech32(addrVal1)
	assert.NilError(t, err)
	delegation, found := f.stakingKeeper.Delegations.Get(ctx, collections.Join(addrAcc, valAddr))
	assert.Assert(t, found)
	var req *types.QueryDelegatorDelegationsRequest

	testCases := []struct {
		msg       string
		malleate  func()
		onSuccess func(response *types.QueryDelegatorDelegationsResponse)
		expErr    bool
		expErrMsg string
	}{
		{
			"empty request",
			func() {
				req = &types.QueryDelegatorDelegationsRequest{}
			},
			func(response *types.QueryDelegatorDelegationsResponse) {},
			true,
			"delegator address cannot be empty",
		},
		{
			"valid request with no delegations",
			func() {
				req = &types.QueryDelegatorDelegationsRequest{DelegatorAddr: addrs[4].String()}
			},
			func(response *types.QueryDelegatorDelegationsResponse) {
				assert.Equal(t, uint64(0), response.Pagination.Total)
				assert.Assert(t, len(response.DelegationResponses) == 0)
			},
			false,
			"",
		},
		{
			"valid request",
			func() {
				req = &types.QueryDelegatorDelegationsRequest{
					DelegatorAddr: addrAcc.String(),
					Pagination:    &query.PageRequest{Limit: 1, CountTotal: true},
				}
			},
			func(response *types.QueryDelegatorDelegationsResponse) {
				assert.Equal(t, uint64(2), response.Pagination.Total)
				assert.Assert(t, len(response.DelegationResponses) == 1)
				assert.DeepEqual(t, sdk.NewCoin(sdk.DefaultBondDenom, delegation.Shares.TruncateInt()), response.DelegationResponses[0].Balance)
			},
			false,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.msg), func(t *testing.T) {
			tc.malleate()
			res, err := queryClient.DelegatorDelegations(gocontext.Background(), req)
			if tc.expErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)
				tc.onSuccess(res)
			}
		})
	}
}

func TestGRPCQueryValidatorDelegations(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx
	addrs, vals := createValidatorAccs(t, f)

	qr := f.app.QueryHelper()
	queryClient := types.NewQueryClient(qr)

	addrAcc := addrs[0]
	addrVal1 := vals[1].OperatorAddress
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrs)
	addrVal2 := valAddrs[4]
	valAddr, err := sdk.ValAddressFromBech32(addrVal1)
	assert.NilError(t, err)
	delegation, found := f.stakingKeeper.Delegations.Get(ctx, collections.Join(addrAcc, valAddr))
	assert.Assert(t, found)

	var req *types.QueryValidatorDelegationsRequest
	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		expErr    bool
		expErrMsg string
	}{
		{
			"empty request",
			func() {
				req = &types.QueryValidatorDelegationsRequest{}
			},
			false,
			true,
			"validator address cannot be empty",
		},
		{
			"invalid validator address",
			func() {
				req = &types.QueryValidatorDelegationsRequest{ValidatorAddr: addrVal2.String()}
			},
			false,
			false,
			"",
		},
		{
			"valid request",
			func() {
				req = &types.QueryValidatorDelegationsRequest{
					ValidatorAddr: addrVal1,
					Pagination:    &query.PageRequest{Limit: 1, CountTotal: true},
				}
			},
			true,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.msg), func(t *testing.T) {
			tc.malleate()
			res, err := queryClient.ValidatorDelegations(gocontext.Background(), req)
			switch {
			case tc.expPass && !tc.expErr:
				assert.NilError(t, err)
				assert.Assert(t, len(res.DelegationResponses) == 1)
				assert.Assert(t, res.Pagination.NextKey != nil)
				assert.Equal(t, uint64(2), res.Pagination.Total)
				assert.Equal(t, addrVal1, res.DelegationResponses[0].Delegation.ValidatorAddress)
				assert.DeepEqual(t, sdk.NewCoin(sdk.DefaultBondDenom, delegation.Shares.TruncateInt()), res.DelegationResponses[0].Balance)
			case !tc.expPass && !tc.expErr:
				assert.NilError(t, err)
				assert.Assert(t, res.DelegationResponses == nil)
			default:
				assert.ErrorContains(t, err, tc.expErrMsg)
				assert.Assert(t, res == nil)
			}
		})
	}
}

func TestGRPCQueryUnbondingDelegation(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx
	addrs, vals := createValidatorAccs(t, f)

	qr := f.app.QueryHelper()
	queryClient := types.NewQueryClient(qr)

	addrAcc2 := addrs[1]
	addrVal2 := vals[1].OperatorAddress

	unbondingTokens := f.stakingKeeper.TokensFromConsensusPower(ctx, 2)
	valAddr, err1 := sdk.ValAddressFromBech32(addrVal2)
	assert.NilError(t, err1)
	_, _, err := f.stakingKeeper.Undelegate(ctx, addrAcc2, valAddr, math.LegacyNewDecFromInt(unbondingTokens))
	assert.NilError(t, err)

	unbond, found := f.stakingKeeper.GetUnbondingDelegation(ctx, addrAcc2, valAddr)
	assert.Assert(t, found)
	var req *types.QueryUnbondingDelegationRequest
	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		expErrMsg string
	}{
		{
			"empty request",
			func() {
				req = &types.QueryUnbondingDelegationRequest{}
			},
			false,
			"delegator address cannot be empty",
		},
		{
			"empty validator address",
			func() {
				req = &types.QueryUnbondingDelegationRequest{
					DelegatorAddr: addrAcc2.String(),
				}
			},
			false,
			"validator address cannot be empty",
		},
		{
			"empty delegator address",
			func() {
				req = &types.QueryUnbondingDelegationRequest{
					ValidatorAddr: addrVal2,
				}
			},
			false,
			"delegator address cannot be empty",
		},
		{
			"invalid validator address",
			func() {
				req = &types.QueryUnbondingDelegationRequest{
					DelegatorAddr: addrAcc2.String(), ValidatorAddr: sdk.AccAddress([]byte("invalid")).String(),
				}
			},
			false,
			"hrp does not match bech32 prefix",
		},
		{
			"delegation not found for validator",
			func() {
				req = &types.QueryUnbondingDelegationRequest{
					DelegatorAddr: addrAcc2.String(), ValidatorAddr: sdk.ValAddress([]byte("invalid")).String(),
				}
			},
			false,
			fmt.Sprintf("unbonding delegation with delegator %s not found for validator", addrAcc2.String()),
		},
		{
			"valid request",
			func() {
				req = &types.QueryUnbondingDelegationRequest{
					DelegatorAddr: addrAcc2.String(), ValidatorAddr: addrVal2,
				}
			},
			true,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.msg), func(t *testing.T) {
			tc.malleate()
			res, err := queryClient.UnbondingDelegation(gocontext.Background(), req)
			if tc.expPass {
				assert.Assert(t, res != nil)
				assert.DeepEqual(t, unbond, res.Unbond)
			} else {
				assert.ErrorContains(t, err, tc.expErrMsg)
				assert.Assert(t, res == nil)
			}
		})
	}
}

func TestGRPCQueryDelegatorUnbondingDelegations(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx
	addrs, vals := createValidatorAccs(t, f)

	qr := f.app.QueryHelper()
	queryClient := types.NewQueryClient(qr)

	addrAcc, addrAcc1 := addrs[0], addrs[1]
	addrVal, addrVal2 := vals[0].OperatorAddress, vals[1].OperatorAddress

	unbondingTokens := f.stakingKeeper.TokensFromConsensusPower(ctx, 2)
	valAddr1, err1 := sdk.ValAddressFromBech32(addrVal)
	assert.NilError(t, err1)
	_, _, err := f.stakingKeeper.Undelegate(ctx, addrAcc, valAddr1, math.LegacyNewDecFromInt(unbondingTokens))
	assert.NilError(t, err)
	valAddr2, err1 := sdk.ValAddressFromBech32(addrVal2)
	assert.NilError(t, err1)
	_, _, err = f.stakingKeeper.Undelegate(ctx, addrAcc, valAddr2, math.LegacyNewDecFromInt(unbondingTokens))
	assert.NilError(t, err)

	unbond, found := f.stakingKeeper.GetUnbondingDelegation(ctx, addrAcc, valAddr1)
	assert.Assert(t, found)
	var req *types.QueryDelegatorUnbondingDelegationsRequest
	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		expErr    bool
		expErrMsg string
	}{
		{
			"empty request",
			func() {
				req = &types.QueryDelegatorUnbondingDelegationsRequest{}
			},
			false,
			true,
			"delegator address cannot be empty",
		},
		{
			"invalid request",
			func() {
				req = &types.QueryDelegatorUnbondingDelegationsRequest{DelegatorAddr: addrAcc1.String()}
			},
			false,
			false,
			"",
		},
		{
			"valid request",
			func() {
				req = &types.QueryDelegatorUnbondingDelegationsRequest{
					DelegatorAddr: addrAcc.String(),
					Pagination:    &query.PageRequest{Limit: 1, CountTotal: true},
				}
			},
			true,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.msg), func(t *testing.T) {
			tc.malleate()
			res, err := queryClient.DelegatorUnbondingDelegations(gocontext.Background(), req)
			switch {
			case tc.expPass && !tc.expErr:
				assert.NilError(t, err)
				assert.Assert(t, res.Pagination.NextKey != nil)
				assert.Equal(t, uint64(2), res.Pagination.Total)
				assert.Assert(t, len(res.UnbondingResponses) == 1)
				assert.DeepEqual(t, unbond, res.UnbondingResponses[0])
			case !tc.expPass && !tc.expErr:
				assert.NilError(t, err)
				assert.Assert(t, res.UnbondingResponses == nil)
			default:
				assert.ErrorContains(t, err, tc.expErrMsg)
				assert.Assert(t, res == nil)
			}
		})
	}
}

func TestGRPCQueryPoolParameters(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx

	qr := f.app.QueryHelper()
	queryClient := types.NewQueryClient(qr)

	bondDenom := sdk.DefaultBondDenom

	// Query pool
	res, err := queryClient.Pool(gocontext.Background(), &types.QueryPoolRequest{})
	assert.NilError(t, err)
	bondedPool := f.stakingKeeper.GetBondedPool(ctx)
	notBondedPool := f.stakingKeeper.GetNotBondedPool(ctx)
	assert.DeepEqual(t, f.bankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), bondDenom).Amount, res.Pool.NotBondedTokens)
	assert.DeepEqual(t, f.bankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount, res.Pool.BondedTokens)

	// Query Params
	resp, err := queryClient.Params(gocontext.Background(), &types.QueryParamsRequest{})
	assert.NilError(t, err)
	params, err := f.stakingKeeper.Params.Get(ctx)
	assert.NilError(t, err)
	assert.DeepEqual(t, params, resp.Params)
}

func TestGRPCQueryRedelegations(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx
	addrs, vals := createValidatorAccs(t, f)

	qr := f.app.QueryHelper()
	queryClient := types.NewQueryClient(qr)

	addrAcc, addrAcc1 := addrs[0], addrs[1]
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrs)
	val1, val2, val3, val4 := vals[0], vals[1], valAddrs[3], valAddrs[4]
	delAmount := f.stakingKeeper.TokensFromConsensusPower(ctx, 1)
	_, err := f.stakingKeeper.Delegate(ctx, addrAcc1, delAmount, types.Unbonded, val1, true)
	assert.NilError(t, err)
	applyValidatorSetUpdates(t, ctx, f.stakingKeeper, -1)

	rdAmount := f.stakingKeeper.TokensFromConsensusPower(ctx, 1)
	val1bz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(val1.GetOperator())
	assert.NilError(t, err)
	val2bz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(val2.GetOperator())
	assert.NilError(t, err)

	_, err = f.stakingKeeper.BeginRedelegation(ctx, addrAcc1, val1bz, val2bz, math.LegacyNewDecFromInt(rdAmount))
	assert.NilError(t, err)
	applyValidatorSetUpdates(t, ctx, f.stakingKeeper, -1)

	redel, found := f.stakingKeeper.Redelegations.Get(ctx, collections.Join3(addrAcc1.Bytes(), valAddrs[0].Bytes(), valAddrs[1].Bytes()))
	assert.Assert(t, found)

	var req *types.QueryRedelegationsRequest
	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		expErr    bool
		expErrMsg string
	}{
		{
			"request redelegations for non existent addr",
			func() {
				req = &types.QueryRedelegationsRequest{DelegatorAddr: addrAcc.String()}
			},
			false,
			false,
			fmt.Sprintf("redelegation not found for delegator address %s", addrAcc.String()),
		},
		{
			"request redelegations with non existent pairs",
			func() {
				req = &types.QueryRedelegationsRequest{
					DelegatorAddr: addrAcc.String(), SrcValidatorAddr: val3.String(),
					DstValidatorAddr: val4.String(),
				}
			},
			false,
			true,
			fmt.Sprintf("redelegation not found for delegator address %s from validator address %s",
				addrAcc.String(), val3.String()),
		},
		{
			"request redelegations with delegatoraddr, sourceValAddr, destValAddr",
			func() {
				req = &types.QueryRedelegationsRequest{
					DelegatorAddr: addrAcc1.String(), SrcValidatorAddr: val1.OperatorAddress,
					DstValidatorAddr: val2.OperatorAddress, Pagination: &query.PageRequest{},
				}
			},
			true,
			false,
			"",
		},
		{
			"request redelegations with delegatoraddr and sourceValAddr",
			func() {
				req = &types.QueryRedelegationsRequest{
					DelegatorAddr: addrAcc1.String(), SrcValidatorAddr: val1.OperatorAddress,
					Pagination: &query.PageRequest{},
				}
			},
			true,
			false,
			"",
		},
		{
			"query redelegations with sourceValAddr only",
			func() {
				req = &types.QueryRedelegationsRequest{
					SrcValidatorAddr: val1.GetOperator(),
					Pagination:       &query.PageRequest{Limit: 1, CountTotal: true},
				}
			},
			true,
			false,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.msg), func(t *testing.T) {
			tc.malleate()
			res, err := queryClient.Redelegations(gocontext.Background(), req)
			switch {
			case tc.expPass && !tc.expErr:
				assert.NilError(t, err)
				assert.Assert(t, len(res.RedelegationResponses) == len(redel.Entries))
				assert.Equal(t, redel.DelegatorAddress, res.RedelegationResponses[0].Redelegation.DelegatorAddress)
				assert.Equal(t, redel.ValidatorSrcAddress, res.RedelegationResponses[0].Redelegation.ValidatorSrcAddress)
				assert.Equal(t, redel.ValidatorDstAddress, res.RedelegationResponses[0].Redelegation.ValidatorDstAddress)
				assert.Assert(t, len(redel.Entries) == len(res.RedelegationResponses[0].Entries))
			case !tc.expPass && !tc.expErr:
				assert.NilError(t, err)
				assert.Assert(t, res.RedelegationResponses == nil)
			default:
				assert.ErrorContains(t, err, tc.expErrMsg)
				assert.Assert(t, res == nil)
			}
		})
	}
}

func TestGRPCQueryValidatorUnbondingDelegations(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx
	addrs, vals := createValidatorAccs(t, f)

	qr := f.app.QueryHelper()
	queryClient := types.NewQueryClient(qr)

	addrAcc1, _ := addrs[0], addrs[1]
	val1 := vals[0]

	// undelegate
	undelAmount := f.stakingKeeper.TokensFromConsensusPower(ctx, 2)
	valbz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(val1.GetOperator())
	assert.NilError(t, err)
	_, _, err = f.stakingKeeper.Undelegate(ctx, addrAcc1, valbz, math.LegacyNewDecFromInt(undelAmount))
	assert.NilError(t, err)
	applyValidatorSetUpdates(t, ctx, f.stakingKeeper, -1)

	var req *types.QueryValidatorUnbondingDelegationsRequest
	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		expErrMsg string
	}{
		{
			"empty request",
			func() {
				req = &types.QueryValidatorUnbondingDelegationsRequest{}
			},
			false,
			"validator address cannot be empty",
		},
		{
			"invalid validator address",
			func() {
				req = &types.QueryValidatorUnbondingDelegationsRequest{
					ValidatorAddr: "invalid",
					Pagination:    &query.PageRequest{Limit: 1, CountTotal: true},
				}
			},
			false,
			"invalid bech32",
		},
		{
			"valid request",
			func() {
				req = &types.QueryValidatorUnbondingDelegationsRequest{
					ValidatorAddr: val1.GetOperator(),
					Pagination:    &query.PageRequest{Limit: 1, CountTotal: true},
				}
			},
			true,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case %s", tc.msg), func(t *testing.T) {
			tc.malleate()
			res, err := queryClient.ValidatorUnbondingDelegations(gocontext.Background(), req)
			if tc.expPass {
				assert.NilError(t, err)
				assert.Equal(t, uint64(1), res.Pagination.Total)
				assert.Equal(t, 1, len(res.UnbondingResponses))
				assert.Equal(t, res.UnbondingResponses[0].ValidatorAddress, val1.OperatorAddress)
			} else {
				assert.ErrorContains(t, err, tc.expErrMsg)
				assert.Assert(t, res == nil)
			}
		})
	}
}
