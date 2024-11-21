package accounts

import (
	"context"
	"testing"

	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/accounts"
	"cosmossdk.io/x/accounts/accountstd"
	account_abstractionv1 "cosmossdk.io/x/accounts/interfaces/account_abstraction/v1"
	accountsv1 "cosmossdk.io/x/accounts/v1"
	"cosmossdk.io/x/bank"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	minttypes "cosmossdk.io/x/mint/types"
	txdecode "cosmossdk.io/x/tx/decode"

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

var _ accountstd.Interface = (*mockAccount)(nil)

type mockAccount struct {
	authenticate func(ctx context.Context, msg *account_abstractionv1.MsgAuthenticate) (*account_abstractionv1.MsgAuthenticateResponse, error)
}

func (m mockAccount) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, func(ctx context.Context, req *gogotypes.Empty) (*gogotypes.Empty, error) {
		return &gogotypes.Empty{}, nil
	})
}

func (m mockAccount) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	if m.authenticate == nil {
		return
	}

	accountstd.RegisterExecuteHandler(builder, m.authenticate)
}

func (m mockAccount) RegisterQueryHandlers(_ *accountstd.QueryBuilder) {}

type fixture struct {
	t *testing.T

	app *integration.App

	cdc codec.Codec
	ctx sdk.Context

	authKeeper     authkeeper.AccountKeeper
	accountsKeeper accounts.Keeper
	bankKeeper     bankkeeper.Keeper

	mockAccountAddress []byte
	bundler            string
}

func (f fixture) mustAddr(address []byte) string {
	s, _ := f.authKeeper.AddressCodec().BytesToString(address)
	return s
}

func (f fixture) runBundle(txBytes ...[]byte) *accountsv1.MsgExecuteBundleResponse {
	f.t.Helper()

	msgSrv := accounts.NewMsgServer(f.accountsKeeper)

	resp, err := msgSrv.ExecuteBundle(f.ctx, &accountsv1.MsgExecuteBundle{
		Bundler: f.bundler,
		Txs:     txBytes,
	})
	require.NoError(f.t, err)
	return resp
}

func (f fixture) mint(address []byte, coins ...sdk.Coin) {
	f.t.Helper()
	for _, coin := range coins {
		err := f.bankKeeper.MintCoins(f.ctx, minttypes.ModuleName, sdk.NewCoins(coin))
		require.NoError(f.t, err)
		err = f.bankKeeper.SendCoinsFromModuleToAccount(f.ctx, minttypes.ModuleName, address, sdk.NewCoins(coin))
		require.NoError(f.t, err)
	}
}

func (f fixture) balance(recipient, denom string) sdk.Coin {
	f.t.Helper()
	balances, err := f.bankKeeper.Balance(f.ctx, &banktypes.QueryBalanceRequest{
		Address: recipient,
		Denom:   denom,
	})
	require.NoError(f.t, err)
	return *balances.Balance
}

func initFixture(t *testing.T, f func(ctx context.Context, msg *account_abstractionv1.MsgAuthenticate) (*account_abstractionv1.MsgAuthenticateResponse, error)) *fixture {
	t.Helper()
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, accounts.StoreKey,
	)
	encodingCfg := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, auth.AppModule{}, bank.AppModule{}, accounts.AppModule{})
	cdc := encodingCfg.Codec

	logger := log.NewTestLogger(t)
	router := baseapp.NewMsgServiceRouter()
	queryRouter := baseapp.NewGRPCQueryRouter()

	txDecoder, err := txdecode.NewDecoder(txdecode.Options{
		SigningContext: encodingCfg.TxConfig.SigningContext(),
		ProtoCodec:     encodingCfg.Codec,
	})
	require.NoError(t, err)

	accountsKeeper, err := accounts.NewKeeper(
		cdc,
		runtime.NewEnvironment(runtime.NewKVStoreService(keys[accounts.StoreKey]), log.NewNopLogger(), runtime.EnvWithQueryRouterService(queryRouter), runtime.EnvWithMsgRouterService(router)),
		addresscodec.NewBech32Codec("cosmos"),
		cdc.InterfaceRegistry(),
		txDecoder,
		accountstd.AddAccount("mock", func(deps accountstd.Dependencies) (accountstd.Interface, error) {
			return mockAccount{f}, nil
		}),
	)
	require.NoError(t, err)
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

	// init account
	_, addr, err := accountsKeeper.Init(integrationApp.Context(), "mock", []byte("system"), &gogotypes.Empty{}, nil)
	require.NoError(t, err)

	fixture := &fixture{
		t:                  t,
		app:                integrationApp,
		cdc:                cdc,
		ctx:                sdk.UnwrapSDKContext(integrationApp.Context()),
		authKeeper:         authKeeper,
		accountsKeeper:     accountsKeeper,
		bankKeeper:         bankKeeper,
		mockAccountAddress: addr,
		bundler:            "",
	}
	fixture.bundler = fixture.mustAddr([]byte("bundler"))
	return fixture
}
