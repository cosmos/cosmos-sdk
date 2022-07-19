package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (suite *IntegrationTestSuite) TestMsgUpdateParams() {
	// default params
	params := types.DefaultParams()

	testCases := []struct {
		name      string
		input     *types.MsgUpdateParams
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid authority",
			input: &types.MsgUpdateParams{
				Authority: "invalid",
				Params:    params,
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "send enabled param",
			input: &types.MsgUpdateParams{
				Authority: suite.app.BankKeeper.GetAuthority(),
				Params: types.Params{
					SendEnabled: []*types.SendEnabled{
						{Denom: "foo", Enabled: true},
					},
				},
			},
			expErr:    true,
			expErrMsg: "use of send_enabled in params is no longer supported",
		},
		{
			name: "all good",
			input: &types.MsgUpdateParams{
				Authority: suite.app.BankKeeper.GetAuthority(),
				Params:    params,
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			_, err := suite.msgServer.UpdateParams(suite.ctx, tc.input)

			if tc.expErr {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}
