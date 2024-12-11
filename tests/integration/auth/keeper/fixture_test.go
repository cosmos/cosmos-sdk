package keeper_test

import (
	"testing"

	cmtabcitypes "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/accounts"
	"cosmossdk.io/x/accounts/accountstd"
	baseaccount "cosmossdk.io/x/accounts/defaults/base"
	accountsv1 "cosmossdk.io/x/accounts/v1"
	"cosmossdk.io/x/bank"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	minttypes "cosmossdk.io/x/mint/types"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/integration"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type fixture struct {
	app *integration.App

	cdc         codec.Codec
	ctx         sdk.Context
	encodingCfg moduletestutil.TestEncodingConfig

	authKeeper     authkeeper.AccountKeeper
	accountsKeeper accounts.Keeper
	bankKeeper     bankkeeper.Keeper
}

func (f fixture) mustAddr(address []byte) string {
	s, _ := f.authKeeper.AddressCodec().BytesToString(address)
	return s
}

func initFixture(t *testing.T, extraAccs map[string]accountstd.Interface) *fixture {
	t.Helper()
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, accounts.StoreKey,
	)
	encodingCfg := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, auth.AppModule{}, bank.AppModule{}, accounts.AppModule{})
	cdc := encodingCfg.Codec

	logger := log.NewTestLogger(t)

	router := baseapp.NewMsgServiceRouter()
	queryRouter := baseapp.NewGRPCQueryRouter()

	handler := directHandler{}
	account := baseaccount.NewAccount("base", signing.NewHandlerMap(handler), baseaccount.WithSecp256K1PubKey())

	var accs []accountstd.AccountCreatorFunc
	for name, acc := range extraAccs {
		f := accountstd.AddAccount(name, func(_ accountstd.Dependencies) (accountstd.Interface, error) {
			return acc, nil
		})
		accs = append(accs, f)
	}
	accountsKeeper, err := accounts.NewKeeper(
		cdc,
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[accounts.StoreKey]), log.NewNopLogger(), runtime.EnvWithQueryRouterService(queryRouter), runtime.EnvWithMsgRouterService(router)),
		addresscodec.NewBech32Codec("cosmos"),
		cdc.InterfaceRegistry(),
		nil,
		append(accs, account)...,
	)
	assert.NilError(t, err)
	accountsv1.RegisterQueryServer(queryRouter, accounts.NewQueryServer(accountsKeeper))

	authority := authtypes.NewModuleAddress("gov")

	authKeeper := authkeeper.NewAccountKeeper(
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[authtypes.StoreKey]), log.NewNopLogger()),
		cdc,
		authtypes.ProtoBaseAccount,
		accountsKeeper,
		map[string][]string{minttypes.ModuleName: {authtypes.Minter}},
		addresscodec.NewBech32Codec(sdk.Bech32MainPrefix),
		sdk.Bech32MainPrefix,
		authority.String(),
	)

	blockedAddresses := map[string]bool{
		authKeeper.GetAuthority(): false,
	}
	bankKeeper := bankkeeper.NewBaseKeeper(
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[banktypes.StoreKey]), log.NewNopLogger()),
		cdc,
		authKeeper,
		blockedAddresses,
		authority.String(),
	)

	accountsModule := accounts.NewAppModule(cdc, accountsKeeper)
	authModule := auth.NewAppModule(cdc, authKeeper, accountsKeeper, authsims.RandomGenesisAccounts, nil)
	bankModule := bank.NewAppModule(cdc, bankKeeper, authKeeper)

	integrationApp := integration.NewIntegrationApp(logger, keys, cdc,
		encodingCfg.InterfaceRegistry.SigningContext().AddressCodec(),
		encodingCfg.InterfaceRegistry.SigningContext().ValidatorAddressCodec(),
		map[string]appmodule.AppModule{
			accounts.ModuleName:  accountsModule,
			authtypes.ModuleName: authModule,
			banktypes.ModuleName: bankModule,
		}, router, queryRouter)

	authtypes.RegisterInterfaces(cdc.InterfaceRegistry())
	banktypes.RegisterInterfaces(cdc.InterfaceRegistry())

	authtypes.RegisterMsgServer(integrationApp.MsgServiceRouter(), authkeeper.NewMsgServerImpl(authKeeper))
	authtypes.RegisterQueryServer(integrationApp.QueryHelper(), authkeeper.NewQueryServer(authKeeper))

	banktypes.RegisterMsgServer(router, bankkeeper.NewMsgServerImpl(bankKeeper))

	// commit and finalize block
	defer func() {
		_, err := integrationApp.Commit()
		if err != nil {
			panic(err)
		}
	}()
	height := integrationApp.LastBlockHeight() + 1
	_, err = integrationApp.FinalizeBlock(&cmtabcitypes.FinalizeBlockRequest{Height: height, DecidedLastCommit: cmtabcitypes.CommitInfo{Votes: []cmtabcitypes.VoteInfo{{}}}})
	require.NoError(t, err)

	return &fixture{
		app:            integrationApp,
		cdc:            cdc,
		ctx:            sdk.UnwrapSDKContext(integrationApp.Context()),
		accountsKeeper: accountsKeeper,
		authKeeper:     authKeeper,
		bankKeeper:     bankKeeper,
		encodingCfg:    encodingCfg,
	}
}
