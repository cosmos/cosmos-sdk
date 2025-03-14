package keeper

import (
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	cmtabcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/protocolpool"
	protocolpoolkeeper "github.com/cosmos/cosmos-sdk/x/protocolpool/keeper"
	protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"testing"
)

var (
	emptyDelAddr sdk.AccAddress
	emptyValAddr sdk.ValAddress

	PKS = simtestutil.CreateTestPubKeys(3)

	valConsPk0 = PKS[0]
)

type fixture struct {
	app *integration.App

	sdkCtx sdk.Context
	cdc    codec.Codec
	keys   map[string]*storetypes.KVStoreKey

	accountKeeper      authkeeper.AccountKeeper
	bankKeeper         bankkeeper.Keeper
	distrKeeper        distrkeeper.Keeper
	stakingKeeper      *stakingkeeper.Keeper
	protocolPoolKeeper protocolpoolkeeper.Keeper

	addr    sdk.AccAddress
	valAddr sdk.ValAddress
}

func initFixture(t testing.TB) *fixture {
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey,
		banktypes.StoreKey,
		distrtypes.StoreKey,
		stakingtypes.StoreKey,
		protocolpooltypes.StoreKey,
	)
	cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, distribution.AppModuleBasic{}).Codec

	logger := log.NewTestLogger(t)
	cms := integration.CreateMultiStore(keys, logger)

	newCtx := sdk.NewContext(cms, types.Header{}, true, logger)

	authority := authtypes.NewModuleAddress("gov")

	maccPerms := map[string][]string{
		distrtypes.ModuleName:                      {authtypes.Minter},
		stakingtypes.BondedPoolName:                {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName:             {authtypes.Burner, authtypes.Staking},
		protocolpooltypes.ModuleName:               {},
		protocolpooltypes.StreamAccount:            {},
		protocolpooltypes.ProtocolPoolDistrAccount: {},
	}

	accountKeeper := authkeeper.NewAccountKeeper(
		cdc,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		addresscodec.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		authority.String(),
	)

	blockedAddresses := map[string]bool{
		accountKeeper.GetAuthority(): false,
	}
	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		accountKeeper,
		blockedAddresses,
		authority.String(),
		log.NewNopLogger(),
	)

	stakingKeeper := stakingkeeper.NewKeeper(
		cdc, runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		accountKeeper,
		bankKeeper,
		authority.String(),
		addresscodec.NewBech32Codec(sdk.Bech32PrefixValAddr),
		addresscodec.NewBech32Codec(sdk.Bech32PrefixConsAddr),
	)

	protocolPoolKeeper := protocolpoolkeeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[protocolpooltypes.ModuleName]),
		accountKeeper,
		bankKeeper,
		authority.String(),
	)

	distrKeeper := distrkeeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[distrtypes.StoreKey]),
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		distrtypes.ModuleName,
		authority.String(),
		distrkeeper.WithExternalCommunityPool(protocolPoolKeeper),
	)

	authModule := auth.NewAppModule(cdc, accountKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil)
	stakingModule := staking.NewAppModule(cdc, stakingKeeper, accountKeeper, bankKeeper, nil)
	distrModule := distribution.NewAppModule(cdc, distrKeeper, accountKeeper, bankKeeper, stakingKeeper, nil)
	protocolPoolModule := protocolpool.NewAppModule(cdc, protocolPoolKeeper, accountKeeper, bankKeeper)

	addr := sdk.AccAddress(PKS[0].Address())
	valAddr := sdk.ValAddress(addr)
	valConsAddr := sdk.ConsAddress(valConsPk0.Address())

	// set proposer and vote infos
	ctx := newCtx.WithProposer(valConsAddr).WithVoteInfos([]cmtabcitypes.VoteInfo{
		{
			Validator: cmtabcitypes.Validator{
				Address: valAddr,
				Power:   100,
			},
			BlockIdFlag: types.BlockIDFlagCommit,
		},
	})

	integrationApp := integration.NewIntegrationApp(ctx, logger, keys, cdc, map[string]appmodule.AppModule{
		authtypes.ModuleName:         authModule,
		banktypes.ModuleName:         bankModule,
		stakingtypes.ModuleName:      stakingModule,
		distrtypes.ModuleName:        distrModule,
		protocolpooltypes.ModuleName: protocolPoolModule,
	})

	sdkCtx := sdk.UnwrapSDKContext(integrationApp.Context())

	// Register MsgServer and QueryServer (x/distribution)
	distrtypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), distrkeeper.NewMsgServerImpl(distrKeeper))
	distrtypes.RegisterQueryServer(integrationApp.QueryHelper(), distrkeeper.NewQuerier(distrKeeper))

	// Register MsgServer and QueryServer (x/protocolpool)
	protocolpooltypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), protocolpoolkeeper.NewMsgServerImpl(protocolPoolKeeper))
	protocolpooltypes.RegisterQueryServer(integrationApp.QueryHelper(), protocolpoolkeeper.NewQuerier(protocolPoolKeeper))

	return &fixture{
		app:                integrationApp,
		sdkCtx:             sdkCtx,
		cdc:                cdc,
		keys:               keys,
		accountKeeper:      accountKeeper,
		bankKeeper:         bankKeeper,
		distrKeeper:        distrKeeper,
		stakingKeeper:      stakingKeeper,
		protocolPoolKeeper: protocolPoolKeeper,
		addr:               addr,
		valAddr:            valAddr,
	}
}
