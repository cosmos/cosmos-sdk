package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func (s *KeeperTestSuite) TestMsgUpdateParams() {
	// default params
	communityTax := sdk.NewDecWithPrec(2, 2)        // 2%
	baseProposerReward := sdk.NewDecWithPrec(1, 2)  // 1%
	bonusProposerReward := sdk.NewDecWithPrec(4, 2) // 4%
	withdrawAddrEnabled := true

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
				Params: types.Params{
					CommunityTax:        sdk.NewDecWithPrec(2, 0),
					BaseProposerReward:  baseProposerReward,
					BonusProposerReward: bonusProposerReward,
					WithdrawAddrEnabled: withdrawAddrEnabled,
				},
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "community tax > 1",
			input: &types.MsgUpdateParams{
				Authority: s.distrKeeper.GetAuthority(),
				Params: types.Params{
					CommunityTax:        sdk.NewDecWithPrec(2, 0),
					BaseProposerReward:  baseProposerReward,
					BonusProposerReward: bonusProposerReward,
					WithdrawAddrEnabled: withdrawAddrEnabled,
				},
			},
			expErr:    true,
			expErrMsg: "community tax should be non-negative and less than one",
		},
		{
			name: "negative community tax",
			input: &types.MsgUpdateParams{
				Authority: s.distrKeeper.GetAuthority(),
				Params: types.Params{
					CommunityTax:        sdk.NewDecWithPrec(-2, 1),
					BaseProposerReward:  baseProposerReward,
					BonusProposerReward: bonusProposerReward,
					WithdrawAddrEnabled: withdrawAddrEnabled,
				},
			},
			expErr:    true,
			expErrMsg: "community tax should be non-negative and less than one",
		},
		{
			name: "base proposer reward > 1",
			input: &types.MsgUpdateParams{
				Authority: s.distrKeeper.GetAuthority(),
				Params: types.Params{
					CommunityTax:        communityTax,
					BaseProposerReward:  sdk.NewDecWithPrec(2, 0),
					BonusProposerReward: bonusProposerReward,
					WithdrawAddrEnabled: withdrawAddrEnabled,
				},
			},
			expErr:    true,
			expErrMsg: "sum of base, bonus proposer rewards, and community tax cannot be greater than one",
		},
		{
			name: "negative base proposer reward",
			input: &types.MsgUpdateParams{
				Authority: s.distrKeeper.GetAuthority(),
				Params: types.Params{
					CommunityTax:        communityTax,
					BaseProposerReward:  sdk.NewDecWithPrec(-2, 0),
					BonusProposerReward: bonusProposerReward,
					WithdrawAddrEnabled: withdrawAddrEnabled,
				},
			},
			expErr:    true,
			expErrMsg: "base proposer reward should be positive",
		},
		{
			name: "bonus proposer reward > 1",
			input: &types.MsgUpdateParams{
				Authority: s.distrKeeper.GetAuthority(),
				Params: types.Params{
					CommunityTax:        communityTax,
					BaseProposerReward:  baseProposerReward,
					BonusProposerReward: sdk.NewDecWithPrec(2, 0),
					WithdrawAddrEnabled: withdrawAddrEnabled,
				},
			},
			expErr:    true,
			expErrMsg: "sum of base, bonus proposer rewards, and community tax cannot be greater than one",
		},
		{
			name: "negative bonus proposer reward",
			input: &types.MsgUpdateParams{
				Authority: s.distrKeeper.GetAuthority(),
				Params: types.Params{
					CommunityTax:        communityTax,
					BaseProposerReward:  baseProposerReward,
					BonusProposerReward: sdk.NewDecWithPrec(-2, 0),
					WithdrawAddrEnabled: withdrawAddrEnabled,
				},
			},
			expErr:    true,
			expErrMsg: "bonus proposer reward should be positive",
		},
		{
			name: "all good",
			input: &types.MsgUpdateParams{
				Authority: s.distrKeeper.GetAuthority(),
				Params: types.Params{
					CommunityTax:        communityTax,
					BaseProposerReward:  baseProposerReward,
					BonusProposerReward: bonusProposerReward,
					WithdrawAddrEnabled: withdrawAddrEnabled,
				},
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			_, err := s.msgServer.UpdateParams(s.ctx, tc.input)

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
