package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtestutil "github.com/cosmos/cosmos-sdk/x/distribution/testutil"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func TestParams(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Height: 1})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		key,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		authtypes.NewModuleAddress("gov").String(),
	)

	// default params
	communityTax := sdkmath.LegacyNewDecWithPrec(2, 2) // 2%
	withdrawAddrEnabled := true

	testCases := []struct {
		name      string
		input     types.Params
		expErr    bool
		expErrMsg string
	}{
		{
			name: "community tax > 1",
			input: types.Params{
				CommunityTax:        sdkmath.LegacyNewDecWithPrec(2, 0),
				BaseProposerReward:  sdkmath.LegacyZeroDec(),
				BonusProposerReward: sdkmath.LegacyZeroDec(),
				WithdrawAddrEnabled: withdrawAddrEnabled,
			},
			expErr:    true,
			expErrMsg: "community tax should be non-negative and less than one",
		},
		{
			name: "negative community tax",
			input: types.Params{
				CommunityTax:        sdkmath.LegacyNewDecWithPrec(-2, 1),
				BaseProposerReward:  sdkmath.LegacyZeroDec(),
				BonusProposerReward: sdkmath.LegacyZeroDec(),
				WithdrawAddrEnabled: withdrawAddrEnabled,
			},
			expErr:    true,
			expErrMsg: "community tax should be non-negative and less than one",
		},
		{
			name: "base proposer reward > 1",
			input: types.Params{
				CommunityTax:        communityTax,
				BaseProposerReward:  sdkmath.LegacyNewDecWithPrec(1, 2),
				BonusProposerReward: sdkmath.LegacyZeroDec(),
				WithdrawAddrEnabled: withdrawAddrEnabled,
			},
			expErr:    false,
			expErrMsg: "base proposer rewards should not be taken into account",
		},
		{
			name: "bonus proposer reward > 1",
			input: types.Params{
				CommunityTax:        communityTax,
				BaseProposerReward:  sdkmath.LegacyNewDecWithPrec(1, 2),
				BonusProposerReward: sdkmath.LegacyZeroDec(),
				WithdrawAddrEnabled: withdrawAddrEnabled,
			},
			expErr:    false,
			expErrMsg: "bonus proposer rewards should not be taken into account",
		},
		{
			name: "all good",
			input: types.Params{
				CommunityTax:        communityTax,
				BaseProposerReward:  sdkmath.LegacyZeroDec(),
				BonusProposerReward: sdkmath.LegacyZeroDec(),
				WithdrawAddrEnabled: withdrawAddrEnabled,
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			expected := distrKeeper.GetParams(ctx)
			err := distrKeeper.SetParams(ctx, tc.input)

			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				expected = tc.input
				require.NoError(t, err)
			}

			params := distrKeeper.GetParams(ctx)
			require.Equal(t, expected, params)
		})
	}
}
