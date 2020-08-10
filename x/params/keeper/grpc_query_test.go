package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
)

func (suite *KeeperTestSuite) TestGRPCQueryParams() {
	var (
		req      *proposal.QueryParamsRequest
		expValue string
		space    types.Subspace
	)
	key := []byte("key")

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = &proposal.QueryParamsRequest{}
			},
			false,
		},
		{
			"invalid request with subspace not found",
			func() {
				req = &proposal.QueryParamsRequest{Subspace: "test"}
			},
			false,
		},
		{
			"invalid request with subspace and key not found",
			func() {
				req = &proposal.QueryParamsRequest{Subspace: "test", Key: "key"}
			},
			false,
		},
		{
			"success",
			func() {
				space = suite.app.ParamsKeeper.Subspace("test").
					WithKeyTable(types.NewKeyTable(types.NewParamSetPair(key, paramJSON{}, validateNoOp)))
				req = &proposal.QueryParamsRequest{Subspace: "test", Key: "key"}
				expValue = ""
			},
			true,
		},
		{
			"update value success",
			func() {
				err := space.Update(suite.ctx, key, []byte(`{"param1":"10241024"}`))
				suite.Require().NoError(err)
				req = &proposal.QueryParamsRequest{Subspace: "test", Key: "key"}
				expValue = `{"param1":"10241024"}`
			},
			true,
		},
	}

	suite.SetupTest()
	ctx := sdk.WrapSDKContext(suite.ctx)

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			tc.malleate()

			res, err := suite.queryClient.Params(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(expValue, res.Param.Value)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(res)
			}
		})
	}
}
