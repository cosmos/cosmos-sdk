package module_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	"github.com/cosmos/cosmos-sdk/x/feegrant/module"
	feegranttestutil "github.com/cosmos/cosmos-sdk/x/feegrant/testutil"
)

func TestFeegrantPruning(t *testing.T) {
	key := storetypes.NewKVStoreKey(feegrant.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModuleBasic{})

	addrs := simtestutil.CreateIncrementalAccounts(4)
	granter1 := addrs[0]
	granter2 := addrs[1]
	granter3 := addrs[2]
	grantee := addrs[3]
	spendLimit := sdk.NewCoins(sdk.NewCoin("stake", sdkmath.NewInt(1000)))
	now := testCtx.Ctx.BlockTime()
	oneDay := now.AddDate(0, 0, 1)

	ctrl := gomock.NewController(t)
	accountKeeper := feegranttestutil.NewMockAccountKeeper(ctrl)
	accountKeeper.EXPECT().GetAccount(gomock.Any(), grantee).Return(authtypes.NewBaseAccountWithAddress(grantee)).AnyTimes()
	accountKeeper.EXPECT().GetAccount(gomock.Any(), granter1).Return(authtypes.NewBaseAccountWithAddress(granter1)).AnyTimes()
	accountKeeper.EXPECT().GetAccount(gomock.Any(), granter2).Return(authtypes.NewBaseAccountWithAddress(granter2)).AnyTimes()
	accountKeeper.EXPECT().GetAccount(gomock.Any(), granter3).Return(authtypes.NewBaseAccountWithAddress(granter3)).AnyTimes()

	accountKeeper.EXPECT().AddressCodec().Return(address.NewBech32Codec("cosmos")).AnyTimes()

	feeGrantKeeper := keeper.NewKeeper(encCfg.Codec, runtime.NewKVStoreService(key), accountKeeper)

	require.NoError(t, feeGrantKeeper.GrantAllowance(
		testCtx.Ctx,
		granter1,
		grantee,
		&feegrant.BasicAllowance{
			Expiration: &now,
		},
	))
	require.NoError(t, feeGrantKeeper.GrantAllowance(
		testCtx.Ctx,
		granter2,
		grantee,
		&feegrant.BasicAllowance{
			SpendLimit: spendLimit,
		},
	))
	require.NoError(t, feeGrantKeeper.GrantAllowance(
		testCtx.Ctx,
		granter3,
		grantee,
		&feegrant.BasicAllowance{
			Expiration: &oneDay,
		},
	))

	queryHelper := baseapp.NewQueryServerTestHelper(testCtx.Ctx, encCfg.InterfaceRegistry)
	feegrant.RegisterQueryServer(queryHelper, feeGrantKeeper)
	queryClient := feegrant.NewQueryClient(queryHelper)

	require.NoError(t, module.EndBlocker(testCtx.Ctx, feeGrantKeeper))

	res, err := queryClient.Allowances(testCtx.Ctx.Context(), &feegrant.QueryAllowancesRequest{
		Grantee: grantee.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res.Allowances, 2)

	testCtx.Ctx = testCtx.Ctx.WithBlockTime(now.AddDate(0, 0, 2))
	require.NoError(t, module.EndBlocker(testCtx.Ctx, feeGrantKeeper))

	res, err = queryClient.Allowances(testCtx.Ctx.Context(), &feegrant.QueryAllowancesRequest{
		Grantee: grantee.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res.Allowances, 1)
}
