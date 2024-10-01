package keeper_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/golang/mock/gomock"
	"gotest.tools/v3/assert"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/bank"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	"cosmossdk.io/x/consensus"
	consensusparamkeeper "cosmossdk.io/x/consensus/keeper"
	consensustypes "cosmossdk.io/x/consensus/types"
	minttypes "cosmossdk.io/x/mint/types"
	pooltypes "cosmossdk.io/x/protocolpool/types"
	"cosmossdk.io/x/staking"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	"cosmossdk.io/x/staking/testutil"
	"cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtestutil "github.com/cosmos/cosmos-sdk/x/auth/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var PKs = simtestutil.CreateTestPubKeys(500)

type fixture struct {
	app *integration.App

	sdkCtx sdk.Context
	cdc    codec.Codec
	keys   map[string]*storetypes.KVStoreKey

	accountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.Keeper
	stakingKeeper *stakingkeeper.Keeper
}

func init() {
	sdk.DefaultPowerReduction = math.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
}

// intended to be used with require/assert:  require.True(ValEq(...))
func ValEq(t *testing.T, exp, got types.Validator) (*testing.T, bool, string, types.Validator, types.Validator) {
	t.Helper()
	return t, exp.MinEqual(&got), "expected:\n%v\ngot:\n%v", exp, got
}

// generateAddresses generates numAddrs of normal AccAddrs and ValAddrs
func generateAddresses(f *fixture, numAddrs int) ([]sdk.AccAddress, []sdk.ValAddress) {
	addrDels := simtestutil.AddTestAddrsIncremental(f.bankKeeper, f.stakingKeeper, f.sdkCtx, numAddrs, math.NewInt(10000))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	return addrDels, addrVals
}

func createValidators(
	t *testing.T,
	f *fixture,
	powers []int64,
) ([]sdk.AccAddress, []sdk.ValAddress, []types.Validator) {
	t.Helper()
	addrs := simtestutil.AddTestAddrsIncremental(f.bankKeeper, f.stakingKeeper, f.sdkCtx, 5, f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 300))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrs)
	pks := simtestutil.CreateTestPubKeys(5)

	val1 := testutil.NewValidator(t, valAddrs[0], pks[0])
	val2 := testutil.NewValidator(t, valAddrs[1], pks[1])
	vals := []types.Validator{val1, val2}

	assert.NilError(t, f.stakingKeeper.SetValidator(f.sdkCtx, val1))
	assert.NilError(t, f.stakingKeeper.SetValidator(f.sdkCtx, val2))
	assert.NilError(t, f.stakingKeeper.SetValidatorByConsAddr(f.sdkCtx, val1))
	assert.NilError(t, f.stakingKeeper.SetValidatorByConsAddr(f.sdkCtx, val2))
	assert.NilError(t, f.stakingKeeper.SetNewValidatorByPowerIndex(f.sdkCtx, val1))
	assert.NilError(t, f.stakingKeeper.SetNewValidatorByPowerIndex(f.sdkCtx, val2))

	for _, addr := range addrs {
		acc := f.accountKeeper.NewAccountWithAddress(f.sdkCtx, addr)
		f.accountKeeper.SetAccount(f.sdkCtx, acc)
	}

	_, err := f.stakingKeeper.Delegate(f.sdkCtx, addrs[0], f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, powers[0]), types.Unbonded, val1, true)
	assert.NilError(t, err)
	_, err = f.stakingKeeper.Delegate(f.sdkCtx, addrs[1], f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, powers[1]), types.Unbonded, val2, true)
	assert.NilError(t, err)
	_, err = f.stakingKeeper.Delegate(f.sdkCtx, addrs[0], f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, powers[2]), types.Unbonded, val2, true)
	assert.NilError(t, err)
	applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, -1)

	return addrs, valAddrs, vals
}

func initFixture(tb testing.TB) *fixture {
	tb.Helper()
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, types.StoreKey, consensustypes.StoreKey,
	)
	encodingCfg := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, auth.AppModule{}, staking.AppModule{})
	cdc := encodingCfg.Codec

	msgRouter := baseapp.NewMsgServiceRouter()
	queryRouter := baseapp.NewGRPCQueryRouter()

	logger := log.NewTestLogger(tb)
	cms := integration.CreateMultiStore(keys, logger)

	newCtx := sdk.NewContext(cms, true, logger)

	authority := authtypes.NewModuleAddress("gov")

	maccPerms := map[string][]string{
		pooltypes.ModuleName:    {},
		minttypes.ModuleName:    {authtypes.Minter},
		types.ModuleName:        {authtypes.Minter},
		types.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		types.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	}

	// gomock initializations
	ctrl := gomock.NewController(tb)
	acctsModKeeper := authtestutil.NewMockAccountsModKeeper(ctrl)

	accountKeeper := authkeeper.NewAccountKeeper(
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[authtypes.StoreKey]), log.NewNopLogger(), runtime.EnvWithQueryRouterService(queryRouter), runtime.EnvWithMsgRouterService(msgRouter)),
		cdc,
		authtypes.ProtoBaseAccount,
		acctsModKeeper,
		maccPerms,
		addresscodec.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		authority.String(),
	)

	blockedAddresses := map[string]bool{
		accountKeeper.GetAuthority(): false,
	}
	bankKeeper := bankkeeper.NewBaseKeeper(
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[banktypes.StoreKey]), log.NewNopLogger()),
		cdc,
		accountKeeper,
		blockedAddresses,
		authority.String(),
	)

	assert.NilError(tb, bankKeeper.SetParams(newCtx, banktypes.DefaultParams()))

	consensusParamsKeeper := consensusparamkeeper.NewKeeper(cdc, runtime.NewEnvironment(runtime.NewKVStoreService(keys[consensustypes.StoreKey]), log.NewNopLogger()), authtypes.NewModuleAddress("gov").String())

	stakingKeeper := stakingkeeper.NewKeeper(cdc, runtime.NewEnvironment(runtime.NewKVStoreService(keys[types.StoreKey]), log.NewNopLogger(), runtime.EnvWithQueryRouterService(queryRouter), runtime.EnvWithMsgRouterService(msgRouter)), accountKeeper, bankKeeper, consensusParamsKeeper, authority.String(), addresscodec.NewBech32Codec(sdk.Bech32PrefixValAddr), addresscodec.NewBech32Codec(sdk.Bech32PrefixConsAddr), runtime.NewContextAwareCometInfoService())

	authModule := auth.NewAppModule(cdc, accountKeeper, acctsModKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper)
	stakingModule := staking.NewAppModule(cdc, stakingKeeper)
	consensusModule := consensus.NewAppModule(cdc, consensusParamsKeeper)

	integrationApp := integration.NewIntegrationApp(newCtx, logger, keys, cdc,
		encodingCfg.InterfaceRegistry.SigningContext().AddressCodec(),
		encodingCfg.InterfaceRegistry.SigningContext().ValidatorAddressCodec(),
		map[string]appmodule.AppModule{
			authtypes.ModuleName:      authModule,
			banktypes.ModuleName:      bankModule,
			types.ModuleName:          stakingModule,
			consensustypes.ModuleName: consensusModule,
		},
		msgRouter,
		queryRouter,
	)

	sdkCtx := sdk.UnwrapSDKContext(integrationApp.Context())

	// Register MsgServer and QueryServer
	types.RegisterMsgServer(integrationApp.MsgServiceRouter(), stakingkeeper.NewMsgServerImpl(stakingKeeper))
	types.RegisterQueryServer(integrationApp.QueryHelper(), stakingkeeper.NewQuerier(stakingKeeper))

	// set default staking params
	assert.NilError(tb, stakingKeeper.Params.Set(sdkCtx, types.DefaultParams()))
	accNum := uint64(0)
	acctsModKeeper.EXPECT().NextAccountNumber(gomock.Any()).AnyTimes().DoAndReturn(func(ctx context.Context) (uint64, error) {
		currentNum := accNum
		accNum++
		return currentNum, nil
	})

	f := fixture{
		app:           integrationApp,
		sdkCtx:        sdkCtx,
		cdc:           cdc,
		keys:          keys,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		stakingKeeper: stakingKeeper,
	}

	return &f
}
