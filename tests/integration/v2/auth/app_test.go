package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/router"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/runtime/v2"
	"cosmossdk.io/runtime/v2/services"
	"cosmossdk.io/server/v2/stf"
	"cosmossdk.io/x/accounts"
	basedepinject "cosmossdk.io/x/accounts/defaults/base/depinject"
	accountsv1 "cosmossdk.io/x/accounts/v1"
	_ "cosmossdk.io/x/bank" // import as blank for app wiring
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	_ "cosmossdk.io/x/consensus" // import as blank for app wiring
	_ "cosmossdk.io/x/staking"   // import as blank for app wirings

	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	_ "github.com/cosmos/cosmos-sdk/x/auth" // import as blank for app wiring
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import as blank for app wiring``
	_ "github.com/cosmos/cosmos-sdk/x/auth/vesting"   // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/genutil"        // import as blank for app wiring
)

type suite struct {
	app *integration.App

	ctx context.Context

	authKeeper     authkeeper.AccountKeeper
	accountsKeeper accounts.Keeper
	bankKeeper     bankkeeper.Keeper
}

func (s suite) mustAddr(address []byte) string {
	str, _ := s.authKeeper.AddressCodec().BytesToString(address)
	return str
}

func createTestSuite(t *testing.T) *suite {
	t.Helper()
	res := suite{}

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

	msgRouterService := integration.NewRouterService()
	res.registerMsgRouterService(msgRouterService)

	var routerFactory runtime.RouterServiceFactory = func(_ []byte) router.Service {
		return msgRouterService
	}

	queryRouterService := integration.NewRouterService()
	res.registerQueryRouterService(queryRouterService)

	serviceBuilder := runtime.NewRouterBuilder(routerFactory, queryRouterService)

	startupCfg.BranchService = &integration.BranchService{}
	startupCfg.RouterServiceBuilder = serviceBuilder
	startupCfg.HeaderService = services.NewGenesisHeaderService(stf.HeaderService{})

	res.app, err = integration.NewApp(
		depinject.Configs(configurator.NewAppV2Config(moduleConfigs...), depinject.Provide(
			// inject desired account types:
			basedepinject.ProvideAccount,

			// provide base account options
			basedepinject.ProvideSecp256K1PubKey,

			// provide extra accounts
			ProvideMockRetroCompatAccountValid,
			ProvideMockRetroCompatAccountNoInfo,
			ProvideMockRetroCompatAccountNoImplement,
		), depinject.Supply(log.NewNopLogger())),
		startupCfg,
		&res.bankKeeper, &res.accountsKeeper, &res.authKeeper)
	require.NoError(t, err)

	res.ctx = res.app.StateLatestContext(t)

	return &res
}

func (s *suite) registerMsgRouterService(router *integration.RouterService) {
	// register custom router service
	bankSendHandler := func(ctx context.Context, req transaction.Msg) (transaction.Msg, error) {
		msg, ok := req.(*banktypes.MsgSend)
		if !ok {
			return nil, integration.ErrInvalidMsgType
		}
		msgServer := bankkeeper.NewMsgServerImpl(s.bankKeeper)
		resp, err := msgServer.Send(ctx, msg)
		return resp, err
	}

	router.RegisterHandler(bankSendHandler, "cosmos.bank.v1beta1.MsgSend")
}

func (s *suite) registerQueryRouterService(router *integration.RouterService) {
	// register custom router service
	queryHandler := func(ctx context.Context, msg transaction.Msg) (transaction.Msg, error) {
		req, ok := msg.(*accountsv1.AccountNumberRequest)
		if !ok {
			return nil, integration.ErrInvalidMsgType
		}
		qs := accounts.NewQueryServer(s.accountsKeeper)
		resp, err := qs.AccountNumber(ctx, req)
		return resp, err
	}

	router.RegisterHandler(queryHandler, "cosmos.accounts.v1.AccountNumberRequest")
}
