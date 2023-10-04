package funds_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	pooltypes "cosmossdk.io/x/protocolpool/types"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/migrations/funds"
	distrtestutil "github.com/cosmos/cosmos-sdk/x/distribution/testutil"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func TestFundsMigration(t *testing.T) {
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, disttypes.StoreKey,
	)
	logger := log.NewTestLogger(t)
	cms := integration.CreateMultiStore(keys, logger)
	encCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{}, distribution.AppModuleBasic{})
	ctx := sdk.NewContext(cms, true, logger)

	maccPerms := map[string][]string{
		pooltypes.ModuleName: nil,
		disttypes.ModuleName: {authtypes.Minter},
	}

	authority := authtypes.NewModuleAddress("gov")

	// create account keeper
	accountKeeper := authkeeper.NewAccountKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		addresscodec.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		authority.String(),
	)

	// create bank keeper
	bankKeeper := bankkeeper.NewBaseKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		accountKeeper,
		map[string]bool{},
		authority.String(),
		log.NewNopLogger(),
	)

	// gomock initializations
	ctrl := gomock.NewController(t)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)
	poolKeeper := distrtestutil.NewMockPoolKeeper(ctrl)

	// create distribution keeper
	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		runtime.NewKVStoreService(keys[disttypes.StoreKey]),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		poolKeeper,
		disttypes.ModuleName,
		authority.String(),
	)

	// Set feepool
	poolAmount := sdk.NewInt64Coin("test", 100000)
	feepool := disttypes.FeePool{
		CommunityPool: sdk.NewDecCoinsFromCoins(poolAmount),
	}
	err := distrKeeper.FeePool.Set(ctx, feepool)
	require.NoError(t, err)

	distrAcc := authtypes.NewEmptyModuleAccount(disttypes.ModuleName)

	// mint coins in distribution module account
	distrModBal := sdk.NewCoins(sdk.NewInt64Coin("test", 10000000))
	err = bankKeeper.MintCoins(ctx, distrAcc.GetName(), distrModBal)
	require.NoError(t, err)

	// Set pool module account
	poolAcc := authtypes.NewEmptyModuleAccount(pooltypes.ModuleName)

	// migrate feepool funds from distribution module account to pool module accout
	err = funds.MigrateFunds(ctx, bankKeeper, feepool, distrAcc, poolAcc)
	require.NoError(t, err)

	// set distrbution feepool as empty (since migration)
	err = distrKeeper.FeePool.Set(ctx, disttypes.FeePool{})
	require.NoError(t, err)

	// check pool module account balance equals pool amount
	poolMAccBal := bankKeeper.GetAllBalances(ctx, poolAcc.GetAddress())
	require.Equal(t, poolMAccBal, sdk.Coins{poolAmount})

	distrAccBal := bankKeeper.GetAllBalances(ctx, distrAcc.GetAddress())
	// check distribution module account balance is not same after migration
	require.NotEqual(t, distrModBal, distrAccBal)
	// check distribution module account balance is same as (current distrAccBal+poolAmount)
	require.Equal(t, distrModBal, distrAccBal.Add(poolAmount))
}
