package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

func (s *KeeperTestSuite) TestUpdateParams() {
	minSignedPerWindow, err := sdk.NewDecFromStr("0.60")
	s.Require().NoError(err)

	slashFractionDoubleSign, err := sdk.NewDecFromStr("0.022")
	s.Require().NoError(err)

	slashFractionDowntime, err := sdk.NewDecFromStr("0.0089")
	s.Require().NoError(err)

	invalidVal, err := sdk.NewDecFromStr("-1")
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		request   *types.MsgUpdateParams
		expectErr bool
		expErrMsg string
	}{
		{
			name: "set invalid authority",
			request: &types.MsgUpdateParams{
				Authority: "foo",
			},
			expectErr: true,
			expErrMsg: "invalid authority",
		},
		{
			name: "set invalid signed blocks window",
			request: &types.MsgUpdateParams{
				Authority: s.slashingKeeper.GetAuthority(),
				Params: types.Params{
					SignedBlocksWindow:      0,
					MinSignedPerWindow:      minSignedPerWindow,
					DowntimeJailDuration:    time.Duration(34800000000000),
					SlashFractionDoubleSign: slashFractionDoubleSign,
					SlashFractionDowntime:   slashFractionDowntime,
				},
			},
			expectErr: true,
			expErrMsg: "signed blocks window must be positive",
		},
		{
			name: "set invalid min signed per window",
			request: &types.MsgUpdateParams{
				Authority: s.slashingKeeper.GetAuthority(),
				Params: types.Params{
					SignedBlocksWindow:      int64(750),
					MinSignedPerWindow:      invalidVal,
					DowntimeJailDuration:    time.Duration(34800000000000),
					SlashFractionDoubleSign: slashFractionDoubleSign,
					SlashFractionDowntime:   slashFractionDowntime,
				},
			},
			expectErr: true,
			expErrMsg: "min signed per window cannot be negative",
		},
		{
			name: "set invalid downtime jail duration",
			request: &types.MsgUpdateParams{
				Authority: s.slashingKeeper.GetAuthority(),
				Params: types.Params{
					SignedBlocksWindow:      int64(750),
					MinSignedPerWindow:      minSignedPerWindow,
					DowntimeJailDuration:    time.Duration(0),
					SlashFractionDoubleSign: slashFractionDoubleSign,
					SlashFractionDowntime:   slashFractionDowntime,
				},
			},
			expectErr: true,
			expErrMsg: "downtime jail duration must be positive",
		},
		{
			name: "set invalid slash fraction double sign",
			request: &types.MsgUpdateParams{
				Authority: s.slashingKeeper.GetAuthority(),
				Params: types.Params{
					SignedBlocksWindow:      int64(750),
					MinSignedPerWindow:      minSignedPerWindow,
					DowntimeJailDuration:    time.Duration(10),
					SlashFractionDoubleSign: invalidVal,
					SlashFractionDowntime:   slashFractionDowntime,
				},
			},
			expectErr: true,
			expErrMsg: "double sign slash fraction cannot be negative",
		},
		{
			name: "set invalid slash fraction downtime",
			request: &types.MsgUpdateParams{
				Authority: s.slashingKeeper.GetAuthority(),
				Params: types.Params{
					SignedBlocksWindow:      int64(750),
					MinSignedPerWindow:      minSignedPerWindow,
					DowntimeJailDuration:    time.Duration(10),
					SlashFractionDoubleSign: slashFractionDoubleSign,
					SlashFractionDowntime:   invalidVal,
				},
			},
			expectErr: true,
			expErrMsg: "downtime slash fraction cannot be negative",
		},
		{
			name: "set full valid params",
			request: &types.MsgUpdateParams{
				Authority: s.slashingKeeper.GetAuthority(),
				Params: types.Params{
					SignedBlocksWindow:      int64(750),
					MinSignedPerWindow:      minSignedPerWindow,
					DowntimeJailDuration:    time.Duration(34800000000000),
					SlashFractionDoubleSign: slashFractionDoubleSign,
					SlashFractionDowntime:   slashFractionDowntime,
				},
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			_, err := s.msgServer.UpdateParams(s.ctx, tc.request)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
