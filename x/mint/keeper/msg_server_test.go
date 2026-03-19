package keeper_test

import (
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

func (s *IntegrationTestSuite) TestUpdateParams() {
	testCases := []struct {
		name      string
		request   *types.MsgUpdateParams
		expectErr bool
	}{
		{
			name: "set invalid authority (not an address)",
			request: &types.MsgUpdateParams{
				Authority: "foo",
			},
			expectErr: true,
		},
		{
			name: "set invalid authority (not defined authority)",
			request: &types.MsgUpdateParams{
				Authority: "cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5",
			},
			expectErr: true,
		},
		{
			name: "set invalid params",
			request: &types.MsgUpdateParams{
				Authority: s.mintKeeper.GetAuthority(),
				Params: types.Params{
					MintDenom:           sdk.DefaultBondDenom,
					InflationRateChange: sdkmath.LegacyNewDecWithPrec(-13, 2),
					InflationMax:        sdkmath.LegacyNewDecWithPrec(20, 2),
					InflationMin:        sdkmath.LegacyNewDecWithPrec(7, 2),
					GoalBonded:          sdkmath.LegacyNewDecWithPrec(67, 2),
					BlocksPerYear:       uint64(60 * 60 * 8766 / 5),
				},
			},
			expectErr: true,
		},
		{
			name: "set full valid params",
			request: &types.MsgUpdateParams{
				Authority: s.mintKeeper.GetAuthority(),
				Params: types.Params{
					MintDenom:           sdk.DefaultBondDenom,
					InflationRateChange: sdkmath.LegacyNewDecWithPrec(8, 2),
					InflationMax:        sdkmath.LegacyNewDecWithPrec(20, 2),
					InflationMin:        sdkmath.LegacyNewDecWithPrec(2, 2),
					GoalBonded:          sdkmath.LegacyNewDecWithPrec(37, 2),
					BlocksPerYear:       uint64(60 * 60 * 8766 / 5),
					MaxSupply:           sdkmath.ZeroInt(), // infinite supply
				},
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := s.msgServer.UpdateParams(s.ctx, tc.request)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestUpdateParamsAuthority() {
	keeperAuthority := s.mintKeeper.GetAuthority()
	overrideAuthority := sdk.AccAddress("override_authority___").String()

	validParams := types.Params{
		MintDenom:           sdk.DefaultBondDenom,
		InflationRateChange: sdkmath.LegacyNewDecWithPrec(8, 2),
		InflationMax:        sdkmath.LegacyNewDecWithPrec(20, 2),
		InflationMin:        sdkmath.LegacyNewDecWithPrec(2, 2),
		GoalBonded:          sdkmath.LegacyNewDecWithPrec(37, 2),
		BlocksPerYear:       uint64(60 * 60 * 8766 / 5),
		MaxSupply:           sdkmath.ZeroInt(),
	}

	s.Run("fallback to keeper authority", func() {
		_, err := s.msgServer.UpdateParams(s.ctx, &types.MsgUpdateParams{
			Authority: keeperAuthority,
			Params:    validParams,
		})
		s.Require().NoError(err)

		_, err = s.msgServer.UpdateParams(s.ctx, &types.MsgUpdateParams{
			Authority: overrideAuthority,
			Params:    validParams,
		})
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "invalid authority")
	})

	s.Run("consensus params authority takes precedence", func() {
		ctx := s.ctx.WithConsensusParams(cmtproto.ConsensusParams{
			Authority: &cmtproto.AuthorityParams{Authority: overrideAuthority},
		})

		_, err := s.msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
			Authority: overrideAuthority,
			Params:    validParams,
		})
		s.Require().NoError(err)

		_, err = s.msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
			Authority: keeperAuthority,
			Params:    validParams,
		})
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "invalid authority")
	})
}
