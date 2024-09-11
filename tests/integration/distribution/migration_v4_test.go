package distribution_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/comet"
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/bank"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	"cosmossdk.io/x/distribution"
	"cosmossdk.io/x/distribution/keeper"
	v4 "cosmossdk.io/x/distribution/migrations/v4"
	distrtestutil "cosmossdk.io/x/distribution/testutil"
	disttypes "cosmossdk.io/x/distribution/types"
	pooltypes "cosmossdk.io/x/protocolpool/types"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtestutil "github.com/cosmos/cosmos-sdk/x/auth/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestMigration(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, distribution.AppModule{}).Codec
	storeKey := storetypes.NewKVStoreKey("distribution")
	storeService := runtime.NewKVStoreService(storeKey)
	tKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)

	env := runtime.NewEnvironment(storeService, coretesting.NewNopLogger())

	addr1 := secp256k1.GenPrivKey().PubKey().Address()
	consAddr1 := sdk.ConsAddress(addr1)

	// Set and check the previous proposer
	err := v4.SetPreviousProposerConsAddr(ctx, storeService, cdc, consAddr1)
	require.NoError(t, err)

	gotAddr, err := v4.GetPreviousProposerConsAddr(ctx, storeService, cdc)
	require.NoError(t, err)
	require.Equal(t, consAddr1, gotAddr)

	err = v4.MigrateStore(ctx, env, cdc)
	require.NoError(t, err)

	// Check that the previous proposer has been removed
	_, err = v4.GetPreviousProposerConsAddr(ctx, storeService, cdc)
	require.ErrorContains(t, err, "previous proposer not set")
}

type emptyCometService struct{}

// CometInfo implements comet.Service.
func (e *emptyCometService) CometInfo(context.Context) comet.Info {
	return comet.Info{}
}

func TestFundsMigration(t *testing.T) {
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, disttypes.StoreKey,
	)
	logger := log.NewTestLogger(t)
	cms := integration.CreateMultiStore(keys, logger)
	encCfg := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, auth.AppModule{}, bank.AppModule{}, distribution.AppModule{})
	ctx := sdk.NewContext(cms, true, logger)
	addressCodec := addresscodec.NewBech32Codec(sdk.Bech32MainPrefix)
	maccPerms := map[string][]string{
		pooltypes.ModuleName: nil,
		disttypes.ModuleName: {authtypes.Minter},
	}

	authority, err := addressCodec.BytesToString(authtypes.NewModuleAddress("gov"))
	require.NoError(t, err)

	// gomock initializations
	ctrl := gomock.NewController(t)
	acctsModKeeper := authtestutil.NewMockAccountsModKeeper(ctrl)
	stakingKeeper := distrtestutil.NewMockStakingKeeper(ctrl)

	accNum := uint64(0)
	acctsModKeeper.EXPECT().NextAccountNumber(gomock.Any()).AnyTimes().DoAndReturn(func(ctx context.Context) (uint64, error) {
		currNum := accNum
		accNum++
		return currNum, nil
	})

	// create account keeper
	accountKeeper := authkeeper.NewAccountKeeper(
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[authtypes.StoreKey]), log.NewNopLogger()),
		encCfg.Codec,
		authtypes.ProtoBaseAccount,
		acctsModKeeper,
		maccPerms,
		addressCodec,
		sdk.Bech32MainPrefix,
		authority,
	)

	// create bank keeper
	bankKeeper := bankkeeper.NewBaseKeeper(
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[banktypes.StoreKey]), log.NewNopLogger()),
		encCfg.Codec,
		accountKeeper,
		map[string]bool{},
		authority,
	)

	// create distribution keeper
	distrKeeper := keeper.NewKeeper(
		encCfg.Codec,
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[disttypes.StoreKey]), log.NewNopLogger()),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		&emptyCometService{},
		disttypes.ModuleName,
		authority,
	)

	// Set feepool
	poolAmount := sdk.NewInt64Coin("test", 100000)
	feepool := disttypes.FeePool{
		CommunityPool: sdk.NewDecCoinsFromCoins(poolAmount),
	}
	err = distrKeeper.FeePool.Set(ctx, feepool)
	require.NoError(t, err)

	distrAcc := authtypes.NewEmptyModuleAccount(disttypes.ModuleName)

	// mint coins in distribution module account
	distrModBal := sdk.NewCoins(sdk.NewInt64Coin("test", 10000000))
	err = bankKeeper.MintCoins(ctx, distrAcc.GetName(), distrModBal)
	require.NoError(t, err)

	// Set pool module account
	poolAcc := authtypes.NewEmptyModuleAccount(pooltypes.ModuleName)

	// migrate feepool funds from distribution module account to pool module account
	_, err = v4.MigrateFunds(ctx, bankKeeper, feepool, distrAcc, poolAcc)
	require.NoError(t, err)

	// set distribution feepool as empty (since migration)
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
