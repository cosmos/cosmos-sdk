package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/x/auth/types"
	banktypes "cosmossdk.io/x/bank/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestAsyncExec() {
	addrs := simtestutil.CreateIncrementalAccounts(2)
	coins := sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(10)))

	msg := &banktypes.MsgSend{
		FromAddress: addrs[0].String(),
		ToAddress:   addrs[1].String(),
		Amount:      coins,
	}

	msgAny, err := codectypes.NewAnyWithValue(msg)
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		req       *types.MsgAsyncExec
		expectErr bool
		expErrMsg string
	}{
		{
			name: "empty signer address",
			req: &types.MsgAsyncExec{
				Signer: "",
				Msgs:   []*codectypes.Any{},
			},
			expectErr: true,
			expErrMsg: "empty signer address string is not allowed",
		},
		{
			name: "invalid signer address",
			req: &types.MsgAsyncExec{
				Signer: "invalid",
				Msgs:   []*codectypes.Any{},
			},
			expectErr: true,
			expErrMsg: "invalid signer address",
		},
		{
			name: "empty msgs",
			req: &types.MsgAsyncExec{
				Signer: addrs[0].String(),
				Msgs:   []*codectypes.Any{},
			},
			expectErr: true,
			expErrMsg: "messages cannot be empty",
		},
		{
			name: "valid msg",
			req: &types.MsgAsyncExec{
				Signer: addrs[0].String(),
				Msgs:   []*codectypes.Any{msgAny},
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			_, err := s.msgServer.AsyncExec(s.ctx, tc.req)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

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
			expErrMsg: "invalid SECK256k1 signature verification cost",
		},
	}

	for _, tc := range testCases {
		tc := tc
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
