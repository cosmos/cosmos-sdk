package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (s *KeeperTestSuite) TestUpdateParams() {
	require := s.Require()

	minSignedPerWindow, err := sdk.NewDecFromStr("0.60")
	require.NoError(err)

	slashFractionDoubleSign, err := sdk.NewDecFromStr("0.022")
	require.NoError(err)

	slashFractionDowntime, err := sdk.NewDecFromStr("0.0089")
	require.NoError(err)

	invalidVal, err := sdk.NewDecFromStr("-1")
	require.NoError(err)

	testCases := []struct {
		name      string
		request   *slashingtypes.MsgUpdateParams
		expectErr bool
		expErrMsg string
	}{
		{
			name: "set invalid authority",
			request: &slashingtypes.MsgUpdateParams{
				Authority: "foo",
			},
			expectErr: true,
			expErrMsg: "invalid authority",
		},
		{
			name: "set invalid signed blocks window",
			request: &slashingtypes.MsgUpdateParams{
				Authority: s.slashingKeeper.GetAuthority(),
				Params: slashingtypes.Params{
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
			request: &slashingtypes.MsgUpdateParams{
				Authority: s.slashingKeeper.GetAuthority(),
				Params: slashingtypes.Params{
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
			request: &slashingtypes.MsgUpdateParams{
				Authority: s.slashingKeeper.GetAuthority(),
				Params: slashingtypes.Params{
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
			request: &slashingtypes.MsgUpdateParams{
				Authority: s.slashingKeeper.GetAuthority(),
				Params: slashingtypes.Params{
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
			request: &slashingtypes.MsgUpdateParams{
				Authority: s.slashingKeeper.GetAuthority(),
				Params: slashingtypes.Params{
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
			request: &slashingtypes.MsgUpdateParams{
				Authority: s.slashingKeeper.GetAuthority(),
				Params: slashingtypes.Params{
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
				require.Error(err)
				require.Contains(err.Error(), tc.expErrMsg)
			} else {
				require.NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestUnjail() {
	testCases := []struct {
		name      string
		malleate  func() *slashingtypes.MsgUnjail
		expErr    bool
		expErrMsg string
	}{
		{
			name: "no self delegation: invalid request",
			malleate: func() *slashingtypes.MsgUnjail {
				_, pubKey, addr := testdata.KeyTestPubAddr()
				valAddr := sdk.ValAddress(addr)
				val, err := types.NewValidator(valAddr, pubKey, types.Description{Moniker: "test"})
				s.Require().NoError(err)

				s.stakingKeeper.EXPECT().Validator(s.ctx, valAddr).Return(val)
				s.stakingKeeper.EXPECT().Delegation(s.ctx, addr, valAddr).Return(nil)

				return &slashingtypes.MsgUnjail{
					ValidatorAddr: sdk.ValAddress(addr).String(),
				}
			},
			expErr:    true,
			expErrMsg: "validator has no self-delegation",
		},
		{
			name: "validator not in the state: invalid request",
			malleate: func() *slashingtypes.MsgUnjail {
				_, _, addr := testdata.KeyTestPubAddr()
				valAddr := sdk.ValAddress(addr)

				s.stakingKeeper.EXPECT().Validator(s.ctx, valAddr).Return(nil)

				return &slashingtypes.MsgUnjail{
					ValidatorAddr: valAddr.String(),
				}
			},
			expErr:    true,
			expErrMsg: "address is not associated with any known validator",
		},
		{
			name: "validator not jailed: invalid request",
			malleate: func() *slashingtypes.MsgUnjail {
				_, pubKey, addr := testdata.KeyTestPubAddr()
				valAddr := sdk.ValAddress(addr)

				val, err := types.NewValidator(valAddr, pubKey, types.Description{Moniker: "test"})
				val.Tokens = sdk.NewInt(1000)
				val.DelegatorShares = sdk.NewDec(1)
				val.Jailed = false

				s.Require().NoError(err)

				info := slashingtypes.NewValidatorSigningInfo(sdk.ConsAddress(addr), int64(4), int64(3),
					time.Unix(2, 0), false, int64(10))

				s.slashingKeeper.SetValidatorSigningInfo(s.ctx, sdk.ConsAddress(addr), info)

				s.stakingKeeper.EXPECT().Validator(s.ctx, valAddr).Return(val)
				del := types.NewDelegation(addr, valAddr, sdk.NewDec(100))

				s.stakingKeeper.EXPECT().Delegation(s.ctx, addr, valAddr).Return(del)

				return &slashingtypes.MsgUnjail{
					ValidatorAddr: sdk.ValAddress(addr).String(),
				}
			},
			expErr:    true,
			expErrMsg: "validator not jailed",
		},
		{
			name: "validator tombstoned: invalid request",
			malleate: func() *slashingtypes.MsgUnjail {
				_, pubKey, addr := testdata.KeyTestPubAddr()
				valAddr := sdk.ValAddress(addr)

				val, err := types.NewValidator(valAddr, pubKey, types.Description{Moniker: "test"})
				val.Tokens = sdk.NewInt(1000)
				val.DelegatorShares = sdk.NewDec(1)
				val.Jailed = true

				s.Require().NoError(err)

				info := slashingtypes.NewValidatorSigningInfo(sdk.ConsAddress(addr), int64(4), int64(3),
					time.Unix(2, 0), true, int64(10))

				s.slashingKeeper.SetValidatorSigningInfo(s.ctx, sdk.ConsAddress(addr), info)

				s.stakingKeeper.EXPECT().Validator(s.ctx, valAddr).Return(val)
				del := types.NewDelegation(addr, valAddr, sdk.NewDec(100))

				s.stakingKeeper.EXPECT().Delegation(s.ctx, addr, valAddr).Return(del)

				return &slashingtypes.MsgUnjail{
					ValidatorAddr: sdk.ValAddress(addr).String(),
				}
			},
			expErr:    true,
			expErrMsg: "validator still jailed; cannot be unjailed",
		},
		{
			name: "unjailing before wait period: invalid request",
			malleate: func() *slashingtypes.MsgUnjail {
				_, pubKey, addr := testdata.KeyTestPubAddr()
				valAddr := sdk.ValAddress(addr)

				val, err := types.NewValidator(valAddr, pubKey, types.Description{Moniker: "test"})
				val.Tokens = sdk.NewInt(1000)
				val.DelegatorShares = sdk.NewDec(1)
				val.Jailed = true

				s.Require().NoError(err)

				info := slashingtypes.NewValidatorSigningInfo(sdk.ConsAddress(addr), int64(4), int64(3),
					s.ctx.BlockTime().AddDate(0, 0, 1), false, int64(10))

				s.slashingKeeper.SetValidatorSigningInfo(s.ctx, sdk.ConsAddress(addr), info)

				s.stakingKeeper.EXPECT().Validator(s.ctx, valAddr).Return(val)
				del := types.NewDelegation(addr, valAddr, sdk.NewDec(10000))

				s.stakingKeeper.EXPECT().Delegation(s.ctx, addr, valAddr).Return(del)

				return &slashingtypes.MsgUnjail{
					ValidatorAddr: sdk.ValAddress(addr).String(),
				}
			},
			expErr:    true,
			expErrMsg: "validator still jailed; cannot be unjailed",
		},
		{
			name: "valid request",
			malleate: func() *slashingtypes.MsgUnjail {
				_, pubKey, addr := testdata.KeyTestPubAddr()
				valAddr := sdk.ValAddress(addr)

				val, err := types.NewValidator(valAddr, pubKey, types.Description{Moniker: "test"})
				val.Tokens = sdk.NewInt(1000)
				val.DelegatorShares = sdk.NewDec(1)

				val.Jailed = true
				s.Require().NoError(err)

				info := slashingtypes.NewValidatorSigningInfo(sdk.ConsAddress(addr), int64(4), int64(3),
					time.Unix(2, 0), false, int64(10))

				s.slashingKeeper.SetValidatorSigningInfo(s.ctx, sdk.ConsAddress(addr), info)

				s.stakingKeeper.EXPECT().Validator(s.ctx, valAddr).Return(val)
				del := types.NewDelegation(addr, valAddr, sdk.NewDec(100))

				s.stakingKeeper.EXPECT().Delegation(s.ctx, addr, valAddr).Return(del)
				s.stakingKeeper.EXPECT().Unjail(s.ctx, sdk.ConsAddress(addr)).Return()

				return &slashingtypes.MsgUnjail{
					ValidatorAddr: sdk.ValAddress(addr).String(),
				}
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			req := tc.malleate()
			_, err := s.msgServer.Unjail(s.ctx, req)

			if tc.expErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
