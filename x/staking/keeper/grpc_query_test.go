package keeper_test

import (
	gocontext "context"
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (s *KeeperTestSuite) TestGRPCQueryValidator() {
	ctx, keeper, queryClient := s.ctx, s.stakingKeeper, s.queryClient
	require := s.Require()

	validator := testutil.NewValidator(s.T(), sdk.ValAddress(PKs[0].Address().Bytes()), PKs[0])
	require.NoError(keeper.SetValidator(ctx, validator))
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
		{
			"with valid and not existing address",
			func() {
				req = &types.QueryValidatorRequest{
					ValidatorAddr: "cosmosvaloper15jkng8hytwt22lllv6mw4k89qkqehtahd84ptu",
				}
			},
			false,
		},
		{
			"valid request",
			func() {
				req = &types.QueryValidatorRequest{ValidatorAddr: validator.OperatorAddress}
			},
			true,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()
			res, err := queryClient.Validator(gocontext.Background(), req)
			if tc.expPass {
				require.NoError(err)
				require.True(validator.Equal(&res.Validator))
			} else {
				require.Error(err)
				require.Nil(res)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGRPCQueryDelegation() {
	s.Run("test delegation query with custom values", func() {
		require := s.Require()
		validatorTokens := "10132421222471611505485287"
		validatorDelegatorShares := "10142563785454078365766009.473933091465315992"
		delegationAmount := "2000000000000000000"

		// Arrange
		ctx, keeper, queryClient := s.ctx, s.stakingKeeper, s.queryClient
		validator := testutil.NewValidator(s.T(), PKs[0].Address().Bytes(), PKs[0])
		tokens, ok := math.NewIntFromString(validatorTokens)
		require.True(ok)
		shares := math.LegacyMustNewDecFromStr(validatorDelegatorShares)
		validator.Tokens = tokens
		validator.DelegatorShares = shares
		require.NoError(keeper.SetValidator(ctx, validator))

		delegator := sdk.AccAddress(PKs[1].Address().Bytes())
		amount, ok := math.NewIntFromString(delegationAmount)
		require.True(ok)
		shares, err := validator.SharesFromTokens(amount)
		require.NoError(err)
		err = keeper.SetDelegation(ctx, types.NewDelegation(delegator.String(), validator.GetOperator(), shares))
		require.NoError(err)

		// Act
		res, err := queryClient.Delegation(gocontext.Background(), &types.QueryDelegationRequest{
			DelegatorAddr: delegator.String(),
			ValidatorAddr: validator.GetOperator(),
		})

		// Assert
		require.NoError(err)
		require.NotNil(res)
		require.True(res.DelegationResponse.Balance.Amount.Equal(amount))
	})
}
