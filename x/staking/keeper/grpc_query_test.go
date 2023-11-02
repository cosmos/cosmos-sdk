package keeper_test

import (
	gocontext "context"
	"fmt"

	"cosmossdk.io/x/staking/testutil"
	"cosmossdk.io/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
