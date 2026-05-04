package keeper_test

import (
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func (s *KeeperTestSuite) TestUpdateParams() {
	testCases := []struct {
		name      string
		req       *types.MsgUpdateParams
		expectErr bool
		expErrMsg string
	}{
		{
			name: "set invalid authority",
			req: &types.MsgUpdateParams{
				Authority: "foo",
			},
			expectErr: true,
			expErrMsg: "invalid authority",
		},
		{
			name: "set invalid max memo characters",
			req: &types.MsgUpdateParams{
				Authority: s.accountKeeper.GetAuthority(),
				Params: types.Params{
					MaxMemoCharacters:      0,
					TxSigLimit:             9,
					TxSizeCostPerByte:      5,
					SigVerifyCostED25519:   694,
					SigVerifyCostSecp256k1: 511,
				},
			},
			expectErr: true,
			expErrMsg: "invalid max memo characters",
		},
		{
			name: "set invalid tx sig limit",
			req: &types.MsgUpdateParams{
				Authority: s.accountKeeper.GetAuthority(),
				Params: types.Params{
					MaxMemoCharacters:      140,
					TxSigLimit:             0,
					TxSizeCostPerByte:      5,
					SigVerifyCostED25519:   694,
					SigVerifyCostSecp256k1: 511,
				},
			},
			expectErr: true,
			expErrMsg: "invalid tx signature limit",
		},
		{
			name: "set invalid tx size cost per bytes",
			req: &types.MsgUpdateParams{
				Authority: s.accountKeeper.GetAuthority(),
				Params: types.Params{
					MaxMemoCharacters:      140,
					TxSigLimit:             9,
					TxSizeCostPerByte:      0,
					SigVerifyCostED25519:   694,
					SigVerifyCostSecp256k1: 511,
				},
			},
			expectErr: true,
			expErrMsg: "invalid tx size cost per byte",
		},
		{
			name: "set invalid sig verify cost ED25519",
			req: &types.MsgUpdateParams{
				Authority: s.accountKeeper.GetAuthority(),
				Params: types.Params{
					MaxMemoCharacters:      140,
					TxSigLimit:             9,
					TxSizeCostPerByte:      5,
					SigVerifyCostED25519:   0,
					SigVerifyCostSecp256k1: 511,
				},
			},
			expectErr: true,
			expErrMsg: "invalid ED25519 signature verification cost",
		},
		{
			name: "set invalid sig verify cost Secp256k1",
			req: &types.MsgUpdateParams{
				Authority: s.accountKeeper.GetAuthority(),
				Params: types.Params{
					MaxMemoCharacters:      140,
					TxSigLimit:             9,
					TxSizeCostPerByte:      5,
					SigVerifyCostED25519:   694,
					SigVerifyCostSecp256k1: 0,
				},
			},
			expectErr: true,
			expErrMsg: "invalid SECP256k1 signature verification cost",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := s.msgServer.UpdateParams(s.ctx, tc.req)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestUpdateParamsAuthority() {
	keeperAuthority := s.accountKeeper.GetAuthority()
	overrideAuthority := sdk.AccAddress("override_authority___").String()

	validParams := types.Params{
		MaxMemoCharacters:      140,
		TxSigLimit:             9,
		TxSizeCostPerByte:      5,
		SigVerifyCostED25519:   694,
		SigVerifyCostSecp256k1: 511,
	}

	s.Run("fallback to keeper authority", func() {
		// No consensus params authority set, keeper authority should work
		_, err := s.msgServer.UpdateParams(s.ctx, &types.MsgUpdateParams{
			Authority: keeperAuthority,
			Params:    validParams,
		})
		s.Require().NoError(err)

		// A different address should fail
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

		// Override authority should now succeed
		_, err := s.msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
			Authority: overrideAuthority,
			Params:    validParams,
		})
		s.Require().NoError(err)

		// Keeper authority should now fail
		_, err = s.msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
			Authority: keeperAuthority,
			Params:    validParams,
		})
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "invalid authority")
	})
}
