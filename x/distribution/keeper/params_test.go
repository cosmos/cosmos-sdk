package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtestutil "github.com/cosmos/cosmos-sdk/x/distribution/testutil"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestParams(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := sdk.NewKVStoreKey(disttypes.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, sdk.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(tmproto.Header{Height: 1})

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
	communityTax := sdk.NewDecWithPrec(2, 2)        // 2%
	baseProposerReward := sdk.NewDecWithPrec(1, 2)  // 1%
	bonusProposerReward := sdk.NewDecWithPrec(4, 2) // 4%
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
				CommunityTax:        sdk.NewDecWithPrec(2, 0),
				BaseProposerReward:  baseProposerReward,
				BonusProposerReward: bonusProposerReward,
				WithdrawAddrEnabled: withdrawAddrEnabled,
			},
			expErr:    true,
			expErrMsg: "community tax should be non-negative and less than one",
		},
		{
			name: "negative community tax",
			input: types.Params{
				CommunityTax:        sdk.NewDecWithPrec(-2, 1),
				BaseProposerReward:  baseProposerReward,
				BonusProposerReward: bonusProposerReward,
				WithdrawAddrEnabled: withdrawAddrEnabled,
			},
			expErr:    true,
			expErrMsg: "community tax should be non-negative and less than one",
		},
		{
			name: "base proposer reward > 1",
			input: types.Params{
				CommunityTax:        communityTax,
				BaseProposerReward:  sdk.NewDecWithPrec(2, 0),
				BonusProposerReward: bonusProposerReward,
				WithdrawAddrEnabled: withdrawAddrEnabled,
			},
			expErr:    true,
			expErrMsg: "sum of base, bonus proposer rewards, and community tax cannot be greater than one",
		},
		{
			name: "negative base proposer reward",
			input: types.Params{
				CommunityTax:        communityTax,
				BaseProposerReward:  sdk.NewDecWithPrec(-2, 0),
				BonusProposerReward: bonusProposerReward,
				WithdrawAddrEnabled: withdrawAddrEnabled,
			},
			expErr:    true,
			expErrMsg: "base proposer reward should be positive",
		},
		{
			name: "bonus proposer reward > 1",
			input: types.Params{
				CommunityTax:        communityTax,
				BaseProposerReward:  baseProposerReward,
				BonusProposerReward: sdk.NewDecWithPrec(2, 0),
				WithdrawAddrEnabled: withdrawAddrEnabled,
			},
			expErr:    true,
			expErrMsg: "sum of base, bonus proposer rewards, and community tax cannot be greater than one",
		},
		{
			name: "negative bonus proposer reward",
			input: types.Params{
				CommunityTax:        communityTax,
				BaseProposerReward:  baseProposerReward,
				BonusProposerReward: sdk.NewDecWithPrec(-2, 0),
				WithdrawAddrEnabled: withdrawAddrEnabled,
			},
			expErr:    true,
			expErrMsg: "bonus proposer reward should be positive",
		},
		{
			name: "all good",
			input: types.Params{
				CommunityTax:        communityTax,
				BaseProposerReward:  baseProposerReward,
				BonusProposerReward: bonusProposerReward,
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
