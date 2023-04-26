package module_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/feegrant"
	"cosmossdk.io/x/feegrant/keeper"
	"cosmossdk.io/x/feegrant/module"
	feegranttestutil "cosmossdk.io/x/feegrant/testutil"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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

	accountKeeper.EXPECT().StringToBytes(grantee.String()).Return(grantee.Bytes(), nil).AnyTimes()
	accountKeeper.EXPECT().StringToBytes(granter1.String()).Return(granter1.Bytes(), nil).AnyTimes()
	accountKeeper.EXPECT().StringToBytes(granter2.String()).Return(granter2.Bytes(), nil).AnyTimes()
	accountKeeper.EXPECT().StringToBytes(granter3.String()).Return(granter3.Bytes(), nil).AnyTimes()

	accountKeeper.EXPECT().BytesToString(grantee.Bytes()).Return(grantee.String(), nil).AnyTimes()
	accountKeeper.EXPECT().BytesToString(granter1.Bytes()).Return(granter1.String(), nil).AnyTimes()
	accountKeeper.EXPECT().BytesToString(granter2.Bytes()).Return(granter2.String(), nil).AnyTimes()
	accountKeeper.EXPECT().BytesToString(granter3.Bytes()).Return(granter3.String(), nil).AnyTimes()

	feegrantKeeper := keeper.NewKeeper(encCfg.Codec, runtime.NewKVStoreService(key), accountKeeper)

	feegrantKeeper.GrantAllowance(
		testCtx.Ctx,
		granter1,
		grantee,
		&feegrant.BasicAllowance{
			Expiration: &now,
		},
	)
	feegrantKeeper.GrantAllowance(
		testCtx.Ctx,
		granter2,
		grantee,
		&feegrant.BasicAllowance{
			SpendLimit: spendLimit,
		},
	)
	feegrantKeeper.GrantAllowance(
		testCtx.Ctx,
		granter3,
		grantee,
		&feegrant.BasicAllowance{
			Expiration: &oneDay,
		},
	)

	queryHelper := baseapp.NewQueryServerTestHelper(testCtx.Ctx, encCfg.InterfaceRegistry)
	feegrant.RegisterQueryServer(queryHelper, feegrantKeeper)
	queryClient := feegrant.NewQueryClient(queryHelper)

	module.EndBlocker(testCtx.Ctx, feegrantKeeper)

	res, err := queryClient.Allowances(testCtx.Ctx.Context(), &feegrant.QueryAllowancesRequest{
		Grantee: grantee.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res.Allowances, 3)

	testCtx.Ctx = testCtx.Ctx.WithBlockTime(now.AddDate(0, 0, 2))
	module.EndBlocker(testCtx.Ctx, feegrantKeeper)

	res, err = queryClient.Allowances(testCtx.Ctx.Context(), &feegrant.QueryAllowancesRequest{
		Grantee: grantee.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res.Allowances, 1)
}
