package keeper_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/distribution/keeper"
	"cosmossdk.io/x/distribution/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestMsgSetWithdrawAddress(t *testing.T) {
	ctx, addrs, distrKeeper, _ := initFixture(t)
	msgServer := keeper.NewMsgServerImpl(distrKeeper)

	cases := []struct {
		name   string
		msg    *types.MsgSetWithdrawAddress
		errMsg string
	}{
		{
			name: "success",
			msg: &types.MsgSetWithdrawAddress{
				DelegatorAddress: addrs[0].String(),
				WithdrawAddress:  addrs[1].String(),
			},
			errMsg: "",
		},
		{
			name: "invalid delegator address",
			msg: &types.MsgSetWithdrawAddress{
				DelegatorAddress: "invalid",
				WithdrawAddress:  addrs[1].String(),
			},
			errMsg: "invalid address",
		},
		{
			name: "invalid withdraw address",
			msg: &types.MsgSetWithdrawAddress{
				DelegatorAddress: addrs[0].String(),
				WithdrawAddress:  "invalid",
			},
			errMsg: "invalid address",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := msgServer.SetWithdrawAddress(ctx, tc.msg)
			if tc.errMsg == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			}
		})
	}
}

func TestMsgWithdrawDelegatorReward(t *testing.T) {
	ctx, addrs, distrKeeper, dep := initFixture(t)
	dep.stakingKeeper.EXPECT().Validator(gomock.Any(), gomock.Any()).AnyTimes()
	msgServer := keeper.NewMsgServerImpl(distrKeeper)

	cases := []struct {
		name   string
		preRun func()
		msg    *types.MsgWithdrawDelegatorReward
		errMsg string
	}{
		{
			name: "invalid delegator address",
			msg: &types.MsgWithdrawDelegatorReward{
				DelegatorAddress: "invalid",
				ValidatorAddress: sdk.ValAddress(addrs[1]).String(),
			},
			errMsg: "invalid delegator address",
		},
		{
			name: "invalid validator address",
			msg: &types.MsgWithdrawDelegatorReward{
				DelegatorAddress: addrs[0].String(),
				ValidatorAddress: "invalid",
			},
			errMsg: "invalid validator address",
		},
		{
			name: "no validator",
			msg: &types.MsgWithdrawDelegatorReward{
				DelegatorAddress: addrs[0].String(),
				ValidatorAddress: sdk.ValAddress(addrs[1]).String(),
			},
			errMsg: "no validator distribution info",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.preRun != nil {
				tc.preRun()
			}
			_, err := msgServer.WithdrawDelegatorReward(ctx, tc.msg)
			if tc.errMsg == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			}
		})
	}
}

func TestMsgWithdrawValidatorCommission(t *testing.T) {
	ctx, addrs, distrKeeper, _ := initFixture(t)
	msgServer := keeper.NewMsgServerImpl(distrKeeper)

	cases := []struct {
		name   string
		preRun func()
		msg    *types.MsgWithdrawValidatorCommission
		errMsg string
	}{
		{
			name: "invalid validator address",
			msg: &types.MsgWithdrawValidatorCommission{
				ValidatorAddress: "invalid",
			},
			errMsg: "invalid validator address",
		},
		{
			name: "no validator commission to withdraw",
			msg: &types.MsgWithdrawValidatorCommission{
				ValidatorAddress: sdk.ValAddress(addrs[1]).String(),
			},
			errMsg: "no validator commission to withdraw",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.preRun != nil {
				tc.preRun()
			}
			_, err := msgServer.WithdrawValidatorCommission(ctx, tc.msg)
			if tc.errMsg == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			}
		})
	}
}

func TestMsgFundCommunityPool(t *testing.T) {
	ctx, addrs, distrKeeper, dep := initFixture(t)
	dep.poolKeeper.EXPECT().FundCommunityPool(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	msgServer := keeper.NewMsgServerImpl(distrKeeper)

	cases := []struct {
		name   string
		msg    *types.MsgFundCommunityPool //nolint:staticcheck // Testing deprecated method
		errMsg string
	}{
		{
			name: "invalid depositor address",
			msg: &types.MsgFundCommunityPool{ //nolint:staticcheck // Testing deprecated method
				Depositor: "invalid",
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
			},
			errMsg: "invalid depositor address",
		},
		{
			name: "success",
			msg: &types.MsgFundCommunityPool{ //nolint:staticcheck // Testing deprecated method
				Depositor: addrs[0].String(),
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(1000))),
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := msgServer.FundCommunityPool(ctx, tc.msg) //nolint:staticcheck // Testing deprecated method
			if tc.errMsg == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			}
		})
	}
}

func TestMsgUpdateParams(t *testing.T) {
	ctx, addrs, distrKeeper, _ := initFixture(t)
	msgServer := keeper.NewMsgServerImpl(distrKeeper)

	cases := []struct {
		name   string
		msg    *types.MsgUpdateParams
		errMsg string
	}{
		{
			name: "invalid authority",
			msg: &types.MsgUpdateParams{
				Authority: "invalid",
				Params:    types.DefaultParams(),
			},
			errMsg: "invalid address",
		},
		{
			name: "incorrect authority",
			msg: &types.MsgUpdateParams{
				Authority: addrs[0].String(),
				Params:    types.DefaultParams(),
			},
			errMsg: "expected authority account as only signer for proposal message",
		},
		{
			name: "invalid params",
			msg: &types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress("gov").String(),
				Params:    types.Params{CommunityTax: math.LegacyNewDec(-1)},
			},
			errMsg: "community tax must be positive",
		},
		{
			name: "success",
			msg: &types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress("gov").String(),
				Params:    types.DefaultParams(),
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := msgServer.UpdateParams(ctx, tc.msg)
			if tc.errMsg == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			}
		})
	}
}

func TestMsgCommunityPoolSpend(t *testing.T) {
	ctx, addrs, distrKeeper, dep := initFixture(t)
	dep.poolKeeper.EXPECT().DistributeFromCommunityPool(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	msgServer := keeper.NewMsgServerImpl(distrKeeper)

	cases := []struct {
		name   string
		msg    *types.MsgCommunityPoolSpend //nolint:staticcheck // Testing deprecated method
		errMsg string
	}{
		{
			name: "invalid authority",
			msg: &types.MsgCommunityPoolSpend{ //nolint:staticcheck // Testing deprecated method
				Authority: "invalid",
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
			},
			errMsg: "invalid address",
		},
		{
			name: "incorrect authority",
			msg: &types.MsgCommunityPoolSpend{ //nolint:staticcheck // Testing deprecated method
				Authority: addrs[0].String(),
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
			},
			errMsg: "expected authority account as only signer for proposal message",
		},
		{
			name: "invalid recipient address",
			msg: &types.MsgCommunityPoolSpend{ //nolint:staticcheck // Testing deprecated method
				Authority: authtypes.NewModuleAddress("gov").String(),
				Recipient: "invalid",
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
			},
			errMsg: "invalid recipient address",
		},
		{
			name: "invalid amount",
			msg: &types.MsgCommunityPoolSpend{ //nolint:staticcheck // Testing deprecated method
				Authority: authtypes.NewModuleAddress("gov").String(),
				Recipient: addrs[0].String(),
			},
			errMsg: "invalid coins",
		},
		{
			name: "success",
			msg: &types.MsgCommunityPoolSpend{ //nolint:staticcheck // Testing deprecated method
				Authority: authtypes.NewModuleAddress("gov").String(),
				Recipient: addrs[0].String(),
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(1000))),
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := msgServer.CommunityPoolSpend(ctx, tc.msg) //nolint:staticcheck // Testing deprecated method
			if tc.errMsg == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			}
		})
	}
}

func TestMsgDepositValidatorRewardsPool(t *testing.T) {
	ctx, _, distrKeeper, _ := initFixture(t)
	msgServer := keeper.NewMsgServerImpl(distrKeeper)

	cases := []struct {
		name   string
		msg    *types.MsgDepositValidatorRewardsPool
		errMsg string
	}{
		{
			name: "invalid depositor address",
			msg: &types.MsgDepositValidatorRewardsPool{
				Depositor: "invalid",
				Amount:    sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100))),
			},
			errMsg: "invalid depositor address",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := msgServer.DepositValidatorRewardsPool(ctx, tc.msg)
			if tc.errMsg == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			}
		})
	}
}
