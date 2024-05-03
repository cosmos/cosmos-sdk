package keeper_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/comet"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/auth"
	authkeeper "cosmossdk.io/x/auth/keeper"
	authsims "cosmossdk.io/x/auth/simulation"
	authtestutil "cosmossdk.io/x/auth/testutil"
	_ "cosmossdk.io/x/auth/tx/config"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/bank"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	"cosmossdk.io/x/distribution"
	distrkeeper "cosmossdk.io/x/distribution/keeper"
	distrtypes "cosmossdk.io/x/distribution/types"
	"cosmossdk.io/x/protocolpool"
	poolkeeper "cosmossdk.io/x/protocolpool/keeper"
	pooltypes "cosmossdk.io/x/protocolpool/types"
	"cosmossdk.io/x/staking"
	_ "cosmossdk.io/x/staking"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	_ "github.com/cosmos/cosmos-sdk/x/consensus"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"

	sdkmath "cosmossdk.io/math"
)

type deterministicFixture struct {
	app *integration.App

	sdkCtx sdk.Context
	cdc    codec.Codec
	keys   map[string]*storetypes.KVStoreKey

	accountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.Keeper
	distrKeeper   distrkeeper.Keeper
	stakingKeeper *stakingkeeper.Keeper
	poolKeeper    poolkeeper.Keeper

	queryClient distrtypes.QueryClient

	addr    sdk.AccAddress
	valAddr sdk.ValAddress
}

func initDeterministicFixture(t *testing.T) *deterministicFixture {
	t.Helper()
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, distrtypes.StoreKey, pooltypes.StoreKey, stakingtypes.StoreKey,
		consensustypes.StoreKey,
	)
	encodingCfg := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, auth.AppModule{}, bank.AppModule{})
	cdc := encodingCfg.Codec

	logger := log.NewTestLogger(t)
	cms := integration.CreateMultiStore(keys, logger)

	newCtx := sdk.NewContext(cms, true, logger)

	authority := authtypes.NewModuleAddress("gov")

	maccPerms := map[string][]string{
		pooltypes.ModuleName:           {},
		pooltypes.StreamAccount:        {},
		distrtypes.ModuleName:          {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	}

	// gomock initializations
	ctrl := gomock.NewController(t)
	acctsModKeeper := authtestutil.NewMockAccountsModKeeper(ctrl)

	accountKeeper := authkeeper.NewAccountKeeper(
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[authtypes.StoreKey]), log.NewNopLogger()),
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

	msgRouter := baseapp.NewMsgServiceRouter()
	grpcRouter := baseapp.NewGRPCQueryRouter()

	stakingKeeper := stakingkeeper.NewKeeper(cdc, runtime.NewEnvironment(runtime.NewKVStoreService(keys[stakingtypes.StoreKey]), log.NewNopLogger(), runtime.EnvWithRouterService(grpcRouter, msgRouter)), accountKeeper, bankKeeper, authority.String(), addresscodec.NewBech32Codec(sdk.Bech32PrefixValAddr), addresscodec.NewBech32Codec(sdk.Bech32PrefixConsAddr))
	require.NoError(t, stakingKeeper.Params.Set(newCtx, stakingtypes.DefaultParams()))

	poolKeeper := poolkeeper.NewKeeper(cdc, runtime.NewEnvironment(runtime.NewKVStoreService(keys[pooltypes.StoreKey]), log.NewNopLogger()), accountKeeper, bankKeeper, stakingKeeper, authority.String())

	distrKeeper := distrkeeper.NewKeeper(
		cdc, runtime.NewEnvironment(runtime.NewKVStoreService(keys[distrtypes.StoreKey]), logger), accountKeeper, bankKeeper, stakingKeeper, poolKeeper, distrtypes.ModuleName, authority.String(),
	)

	authModule := auth.NewAppModule(cdc, accountKeeper, acctsModKeeper, authsims.RandomGenesisAccounts)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper)
	stakingModule := staking.NewAppModule(cdc, stakingKeeper, accountKeeper, bankKeeper)
	distrModule := distribution.NewAppModule(cdc, distrKeeper, accountKeeper, bankKeeper, stakingKeeper, poolKeeper)
	poolModule := protocolpool.NewAppModule(cdc, poolKeeper, accountKeeper, bankKeeper)

	addr := sdk.AccAddress(PKS[0].Address())
	valAddr := sdk.ValAddress(addr)
	valConsAddr := sdk.ConsAddress(valConsPk0.Address())

	// set proposer and vote infos
	ctx := newCtx.WithProposer(valConsAddr).WithCometInfo(comet.Info{
		LastCommit: comet.CommitInfo{
			Votes: []comet.VoteInfo{
				{
					Validator: comet.Validator{
						Address: valAddr,
						Power:   100,
					},
					BlockIDFlag: comet.BlockIDFlagCommit,
				},
			},
		},
	})

	integrationApp := integration.NewIntegrationApp(ctx, logger, keys, cdc,
		encodingCfg.InterfaceRegistry.SigningContext().AddressCodec(),
		encodingCfg.InterfaceRegistry.SigningContext().ValidatorAddressCodec(),
		map[string]appmodule.AppModule{
			authtypes.ModuleName:    authModule,
			banktypes.ModuleName:    bankModule,
			stakingtypes.ModuleName: stakingModule,
			distrtypes.ModuleName:   distrModule,
			pooltypes.ModuleName:    poolModule,
		},
		msgRouter,
		grpcRouter,
	)

	sdkCtx := sdk.UnwrapSDKContext(integrationApp.Context())

	// Register MsgServer and QueryServer
	distrtypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), distrkeeper.NewMsgServerImpl(distrKeeper))
	distrtypes.RegisterQueryServer(integrationApp.QueryHelper(), distrkeeper.NewQuerier(distrKeeper))

	qr := integrationApp.QueryHelper()
	queryClient := distrtypes.NewQueryClient(qr)

	return &deterministicFixture{
		app:           integrationApp,
		sdkCtx:        sdkCtx,
		cdc:           cdc,
		keys:          keys,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		distrKeeper:   distrKeeper,
		stakingKeeper: stakingKeeper,
		poolKeeper:    poolKeeper,
		addr:          addr,
		valAddr:       valAddr,
		queryClient:   queryClient,
	}
}

func TestQueryParamsDeterministic(t *testing.T) {
	t.Parallel()
	f := initDeterministicFixture(t)

	rapid.Check(t, func(rt *rapid.T) {
		req := &distrtypes.QueryParamsRequest{}

		communityTaxGen := rapid.Map(rapid.Float64Range(0.0, 0.2), func(t float64) sdkmath.LegacyDec {
			return sdkmath.LegacyNewDecWithPrec(int64(t*100), 2)
		})
		baseProposerRewardGen := rapid.Map(rapid.Float64Range(0.0, 0.1), func(t float64) sdkmath.LegacyDec {
			return sdkmath.LegacyNewDecWithPrec(int64(t*100), 2)
		})
		bonusProposerRewardGen := rapid.Map(rapid.Float64Range(0.0, 0.1), func(t float64) sdkmath.LegacyDec {
			return sdkmath.LegacyNewDecWithPrec(int64(t*100), 2)
		})
		withdrawAddrEnabledGen := rapid.Bool()

		// set the params
		err := f.distrKeeper.Params.Set(f.sdkCtx, distrtypes.Params{
			CommunityTax:        communityTaxGen.Draw(rt, "community_tax"),
			BaseProposerReward:  baseProposerRewardGen.Draw(rt, "base_proposer_reward"),
			BonusProposerReward: bonusProposerRewardGen.Draw(rt, "bonus_proposer_reward"),
			WithdrawAddrEnabled: withdrawAddrEnabledGen.Draw(rt, "withdraw_addr_enabled"),
		})

		if err != nil {
			rt.Fatalf("error setting params: %v", err)
		}

		testdata.DeterministicIterations(t, f.sdkCtx, req, f.queryClient.Params, 0, true)
	})
}
