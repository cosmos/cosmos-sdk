package keeper_test

import (
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtestutil "github.com/cosmos/cosmos-sdk/x/distribution/testutil"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func TestUpdateParamsAuthority(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: time.Now()})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

	keeperAuthority := authtypes.NewModuleAddress("gov").String()
	overrideAuthority := sdk.AccAddress("override_authority___").String()

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		keeperAuthority,
	)

	require.NoError(t, distrKeeper.Params.Set(ctx, types.DefaultParams()))
	msgServer := keeper.NewMsgServerImpl(distrKeeper)

	t.Run("fallback to keeper authority", func(t *testing.T) {
		_, err := msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
			Authority: keeperAuthority,
			Params:    types.DefaultParams(),
		})
		require.NoError(t, err)

		_, err = msgServer.UpdateParams(ctx, &types.MsgUpdateParams{
			Authority: overrideAuthority,
			Params:    types.DefaultParams(),
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid authority")
	})

	t.Run("consensus params authority takes precedence", func(t *testing.T) {
		ctxOverride := ctx.WithConsensusParams(cmtproto.ConsensusParams{
			Authority: &cmtproto.AuthorityParams{Authority: overrideAuthority},
		})

		_, err := msgServer.UpdateParams(ctxOverride, &types.MsgUpdateParams{
			Authority: overrideAuthority,
			Params:    types.DefaultParams(),
		})
		require.NoError(t, err)

		_, err = msgServer.UpdateParams(ctxOverride, &types.MsgUpdateParams{
			Authority: keeperAuthority,
			Params:    types.DefaultParams(),
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid authority")
	})
}

func TestCommunityPoolSpendAuthority(t *testing.T) {
	ctrl := gomock.NewController(t)
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{})
	ctx := testCtx.Ctx.WithBlockHeader(cmtproto.Header{Time: time.Now()})

	bankKeeper := distrtestutil.NewMockBankKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	accountKeeper := distrtestutil.NewMockAccountKeeper(ctrl)

	accountKeeper.EXPECT().GetModuleAddress("distribution").Return(distrAcc.GetAddress())
	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

	keeperAuthority := authtypes.NewModuleAddress("gov").String()
	overrideAuthority := sdk.AccAddress("override_authority___").String()

	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		"fee_collector",
		keeperAuthority,
	)

	require.NoError(t, distrKeeper.Params.Set(ctx, types.DefaultParams()))
	msgServer := keeper.NewMsgServerImpl(distrKeeper)

	amount := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100)))
	recipient := sdk.AccAddress("recipient____________").String()

	t.Run("fallback to keeper authority", func(t *testing.T) {
		// Keeper authority should pass authority check (may fail for other reasons)
		_, err := msgServer.CommunityPoolSpend(ctx, &types.MsgCommunityPoolSpend{
			Authority: overrideAuthority,
			Recipient: recipient,
			Amount:    amount,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid authority")
	})

	t.Run("consensus params authority takes precedence", func(t *testing.T) {
		ctxOverride := ctx.WithConsensusParams(cmtproto.ConsensusParams{
			Authority: &cmtproto.AuthorityParams{Authority: overrideAuthority},
		})

		// Keeper authority should now fail
		_, err := msgServer.CommunityPoolSpend(ctxOverride, &types.MsgCommunityPoolSpend{
			Authority: keeperAuthority,
			Recipient: recipient,
			Amount:    amount,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid authority")
	})
}
