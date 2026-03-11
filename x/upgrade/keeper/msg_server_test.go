package keeper_test

import (
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func (s *KeeperTestSuite) TestSoftwareUpgrade() {
	govAccAddr := sdk.AccAddress(address.Module("gov")).String()

	testCases := []struct {
		name      string
		req       *types.MsgSoftwareUpgrade
		expectErr bool
		errMsg    string
	}{
		{
			"invalid authority address",
			&types.MsgSoftwareUpgrade{
				Authority: "authority",
				Plan: types.Plan{
					Name:   "all-good",
					Height: 123450000,
				},
			},
			true,
			"invalid authority",
		},
		{
			"unauthorized authority address",
			&types.MsgSoftwareUpgrade{
				Authority: s.addrs[0].String(),
				Plan: types.Plan{
					Name:   "all-good",
					Info:   "some text here",
					Height: 123450000,
				},
			},
			true,
			"invalid authority",
		},
		{
			"invalid plan",
			&types.MsgSoftwareUpgrade{
				Authority: govAccAddr,
				Plan: types.Plan{
					Height: 123450000,
				},
			},
			true,
			"name cannot be empty: invalid request",
		},
		{
			"successful upgrade scheduled",
			&types.MsgSoftwareUpgrade{
				Authority: govAccAddr,
				Plan: types.Plan{
					Name:   "all-good",
					Info:   "some text here",
					Height: 123450000,
				},
			},
			false,
			"",
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := s.msgSrvr.SoftwareUpgrade(s.ctx, tc.req)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errMsg)
			} else {
				s.Require().NoError(err)
				plan, err := s.upgradeKeeper.GetUpgradePlan(s.ctx)
				s.Require().NoError(err)
				s.Require().Equal(tc.req.Plan, plan)
			}
		})
	}
}

func (s *KeeperTestSuite) TestCancelUpgrade() {
	govAccAddr := "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn" // TODO
	// govAccAddr := s.govKeeper.GetGovernanceAccount(s.ctx).GetAddress().String()
	err := s.upgradeKeeper.ScheduleUpgrade(s.ctx, types.Plan{
		Name:   "some name",
		Info:   "some info",
		Height: 123450000,
	})
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		req       *types.MsgCancelUpgrade
		expectErr bool
		errMsg    string
	}{
		{
			"invalid authority address",
			&types.MsgCancelUpgrade{
				Authority: "authority",
			},
			true,
			"invalid authority",
		},
		{
			"unauthorized authority address",
			&types.MsgCancelUpgrade{
				Authority: s.addrs[0].String(),
			},
			true,
			"invalid authority",
		},
		{
			"upgrade canceled successfully",
			&types.MsgCancelUpgrade{
				Authority: govAccAddr,
			},
			false,
			"",
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := s.msgSrvr.CancelUpgrade(s.ctx, tc.req)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errMsg)
			} else {
				s.Require().NoError(err)
				_, err := s.upgradeKeeper.GetUpgradePlan(s.ctx)
				s.Require().ErrorIs(err, types.ErrNoUpgradePlanFound)
			}
		})
	}
}

func (s *KeeperTestSuite) TestSoftwareUpgradeAuthority() {
	keeperAuthority := sdk.AccAddress(address.Module("gov")).String()
	overrideAuthority := sdk.AccAddress("override_authority___").String()

	validPlan := types.Plan{
		Name:   "authority-test",
		Info:   "test info",
		Height: 123450000,
	}

	s.Run("fallback to keeper authority", func() {
		_, err := s.msgSrvr.SoftwareUpgrade(s.ctx, &types.MsgSoftwareUpgrade{
			Authority: keeperAuthority,
			Plan:      validPlan,
		})
		s.Require().NoError(err)

		_, err = s.msgSrvr.SoftwareUpgrade(s.ctx, &types.MsgSoftwareUpgrade{
			Authority: overrideAuthority,
			Plan:      validPlan,
		})
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "expected")
	})

	s.Run("consensus params authority takes precedence", func() {
		ctx := s.ctx.WithConsensusParams(cmtproto.ConsensusParams{
			Authority: &cmtproto.AuthorityParams{Authority: overrideAuthority},
		})

		_, err := s.msgSrvr.SoftwareUpgrade(ctx, &types.MsgSoftwareUpgrade{
			Authority: overrideAuthority,
			Plan: types.Plan{
				Name:   "authority-test-override",
				Info:   "test info",
				Height: 123450001,
			},
		})
		s.Require().NoError(err)

		_, err = s.msgSrvr.SoftwareUpgrade(ctx, &types.MsgSoftwareUpgrade{
			Authority: keeperAuthority,
			Plan:      validPlan,
		})
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "expected")
	})
}

func (s *KeeperTestSuite) TestCancelUpgradeAuthority() {
	keeperAuthority := sdk.AccAddress(address.Module("gov")).String()
	overrideAuthority := sdk.AccAddress("override_authority___").String()

	// Schedule an upgrade first
	err := s.upgradeKeeper.ScheduleUpgrade(s.ctx, types.Plan{
		Name:   "cancel-auth-test",
		Info:   "some info",
		Height: 123450000,
	})
	s.Require().NoError(err)

	s.Run("fallback to keeper authority", func() {
		_, err := s.msgSrvr.CancelUpgrade(s.ctx, &types.MsgCancelUpgrade{
			Authority: overrideAuthority,
		})
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "expected")

		_, err = s.msgSrvr.CancelUpgrade(s.ctx, &types.MsgCancelUpgrade{
			Authority: keeperAuthority,
		})
		s.Require().NoError(err)
	})

	s.Run("consensus params authority takes precedence", func() {
		// Re-schedule
		err := s.upgradeKeeper.ScheduleUpgrade(s.ctx, types.Plan{
			Name:   "cancel-auth-test-2",
			Info:   "some info",
			Height: 123450000,
		})
		s.Require().NoError(err)

		ctx := s.ctx.WithConsensusParams(cmtproto.ConsensusParams{
			Authority: &cmtproto.AuthorityParams{Authority: overrideAuthority},
		})

		_, err = s.msgSrvr.CancelUpgrade(ctx, &types.MsgCancelUpgrade{
			Authority: keeperAuthority,
		})
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "expected")

		_, err = s.msgSrvr.CancelUpgrade(ctx, &types.MsgCancelUpgrade{
			Authority: overrideAuthority,
		})
		s.Require().NoError(err)
	})
}
