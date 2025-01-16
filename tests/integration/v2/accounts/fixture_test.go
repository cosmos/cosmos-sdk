package accounts

import (
	"context"
	"testing"

	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/router"
	"cosmossdk.io/core/testing/msgrouter"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/runtime/v2"
	"cosmossdk.io/x/accounts"
	"cosmossdk.io/x/accounts/accountstd"
	basedepinject "cosmossdk.io/x/accounts/defaults/base/depinject"
	account_abstractionv1 "cosmossdk.io/x/accounts/interfaces/account_abstraction/v1"
	counteraccount "cosmossdk.io/x/accounts/testing/counter"
	accountsv1 "cosmossdk.io/x/accounts/v1"
	"cosmossdk.io/x/bank"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	_ "cosmossdk.io/x/consensus" // import as blank for app wiring
	minttypes "cosmossdk.io/x/mint/types"
	_ "cosmossdk.io/x/staking" // import as blank for app wirings

	"github.com/cosmos/cosmos-sdk/codec"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import as blank for app wiring``
	_ "github.com/cosmos/cosmos-sdk/x/auth/vesting"   // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/genutil"        // import as blank for app wiring
)

var _ accountstd.Interface = (*mockAccount)(nil)

type mockAccount struct {
	authenticate authentiacteFunc
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

func ProvideMockAccount(f authentiacteFunc) accountstd.DepinjectAccount {
	return accountstd.DepinjectAccount{MakeAccount: func(_ accountstd.Dependencies) (string, accountstd.Interface, error) {
		return "mock", mockAccount{f}, nil
	}}
}

type authentiacteFunc = func(ctx context.Context, msg *account_abstractionv1.MsgAuthenticate) (*account_abstractionv1.MsgAuthenticateResponse, error)

type fixture struct {
	t *testing.T

	app *integration.App
	cdc codec.Codec
	ctx context.Context

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

func initFixture(t *testing.T, f authentiacteFunc) *fixture {
	t.Helper()

	fixture := &fixture{}
	fixture.t = t
	encodingCfg := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, auth.AppModule{}, bank.AppModule{}, accounts.AppModule{})
	cdc := encodingCfg.Codec

	moduleConfigs := []configurator.ModuleOption{
		configurator.AccountsModule(),
		configurator.AuthModule(),
		configurator.BankModule(),
		configurator.VestingModule(),
		configurator.StakingModule(),
		configurator.TxModule(),
		configurator.ValidateModule(),
		configurator.ConsensusModule(),
		configurator.GenutilModule(),
	}

	var err error
	startupCfg := integration.DefaultStartUpConfig(t)

	msgRouterService := msgrouter.NewRouterService()
	fixture.registerMsgRouterService(msgRouterService)

	var routerFactory runtime.RouterServiceFactory = func(_ []byte) router.Service {
		return msgRouterService
	}

	queryRouterService := msgrouter.NewRouterService()
	fixture.registerQueryRouterService(queryRouterService)

	serviceBuilder := runtime.NewRouterBuilder(routerFactory, queryRouterService)

	startupCfg.BranchService = &integration.BranchService{}
	startupCfg.RouterServiceBuilder = serviceBuilder
	startupCfg.HeaderService = &integration.HeaderService{}
	startupCfg.GasService = &integration.GasService{}

	fixture.app, err = integration.NewApp(
		depinject.Configs(configurator.NewAppV2Config(moduleConfigs...), depinject.Provide(
			// inject desired account types:
			basedepinject.ProvideAccount,

			// provide base account options
			basedepinject.ProvideSecp256K1PubKey,

			ProvideMockAccount,
			counteraccount.ProvideAccount,
		), depinject.Supply(log.NewNopLogger(), f)),
		startupCfg,
		&fixture.bankKeeper, &fixture.accountsKeeper, &fixture.authKeeper, &fixture.cdc)
	require.NoError(t, err)

	fixture.ctx = fixture.app.StateLatestContext(t)

	// init account
	_, addr, err := fixture.accountsKeeper.Init(fixture.ctx, "mock", []byte("system"), &gogotypes.Empty{}, nil, nil)
	require.NoError(t, err)

	fixture.cdc = cdc
	fixture.mockAccountAddress = addr
	fixture.bundler = fixture.mustAddr([]byte("bundler"))
	return fixture
}

func (f *fixture) registerMsgRouterService(router *msgrouter.RouterService) {
	// register custom router service
	bankSendHandler := func(ctx context.Context, req transaction.Msg) (transaction.Msg, error) {
		msg, ok := req.(*banktypes.MsgSend)
		if !ok {
			return nil, integration.ErrInvalidMsgType
		}
		msgServer := bankkeeper.NewMsgServerImpl(f.bankKeeper)
		resp, err := msgServer.Send(ctx, msg)
		return resp, err
	}

	router.RegisterHandler(bankSendHandler, "cosmos.bank.v1beta1.MsgSend")
}

func (f *fixture) registerQueryRouterService(router *msgrouter.RouterService) {
	// register custom router service
	queryHandler := func(ctx context.Context, msg transaction.Msg) (transaction.Msg, error) {
		req, ok := msg.(*accountsv1.AccountNumberRequest)
		if !ok {
			return nil, integration.ErrInvalidMsgType
		}
		qs := accounts.NewQueryServer(f.accountsKeeper)
		resp, err := qs.AccountNumber(ctx, req)
		return resp, err
	}

	router.RegisterHandler(queryHandler, "cosmos.accounts.v1.AccountNumberRequest")
}
