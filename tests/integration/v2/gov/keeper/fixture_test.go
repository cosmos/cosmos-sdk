package keeper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/router"
	"cosmossdk.io/core/testing/msgrouter"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/runtime/v2"
	_ "cosmossdk.io/x/accounts"
	_ "cosmossdk.io/x/bank"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	_ "cosmossdk.io/x/consensus"
	_ "cosmossdk.io/x/gov"
	govkeeper "cosmossdk.io/x/gov/keeper"
	v1 "cosmossdk.io/x/gov/types/v1"
	"cosmossdk.io/x/gov/types/v1beta1"
	_ "cosmossdk.io/x/protocolpool"
	_ "cosmossdk.io/x/staking"
	stakingkeeper "cosmossdk.io/x/staking/keeper"

	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	_ "github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
)

type fixture struct {
	ctx context.Context
	app *integration.App

	queryServer       v1.QueryServer
	legacyQueryServer v1beta1.QueryServer

	authKeeper    authkeeper.AccountKeeper
	bankKeeper    bankkeeper.Keeper
	stakingKeeper *stakingkeeper.Keeper
	govKeeper     *govkeeper.Keeper
}

func initFixture(t *testing.T) *fixture {
	t.Helper()
	res := fixture{}

	moduleConfigs := []configurator.ModuleOption{
		configurator.AccountsModule(),
		configurator.AuthModule(),
		configurator.StakingModule(),
		configurator.BankModule(),
		configurator.TxModule(),
		configurator.GovModule(),
		configurator.ConsensusModule(),
		configurator.ProtocolPoolModule(),
	}

	startupCfg := integration.DefaultStartUpConfig(t)

	msgRouterService := msgrouter.NewRouterService()
	res.registerMsgRouterService(msgRouterService)

	var routerFactory runtime.RouterServiceFactory = func(_ []byte) router.Service {
		return msgRouterService
	}

	queryRouterService := msgrouter.NewRouterService()
	res.registerQueryRouterService(queryRouterService)
	serviceBuilder := runtime.NewRouterBuilder(routerFactory, queryRouterService)

	startupCfg.BranchService = &integration.BranchService{}
	startupCfg.RouterServiceBuilder = serviceBuilder
	startupCfg.HeaderService = &integration.HeaderService{}

	app, err := integration.NewApp(
		depinject.Configs(configurator.NewAppV2Config(moduleConfigs...), depinject.Supply(log.NewNopLogger())),
		startupCfg,
		&res.authKeeper, &res.bankKeeper, &res.govKeeper, &res.stakingKeeper)
	require.NoError(t, err)

	res.app = app
	res.ctx = app.StateLatestContext(t)

	res.queryServer = govkeeper.NewQueryServer(res.govKeeper)
	res.legacyQueryServer = govkeeper.NewLegacyQueryServer(res.govKeeper)
	return &res
}

func (f *fixture) registerMsgRouterService(router *msgrouter.RouterService) {
	// register custom router service

	govSubmitProposalHandler := func(ctx context.Context, req transaction.Msg) (transaction.Msg, error) {
		msg, ok := req.(*v1.MsgExecLegacyContent)
		if !ok {
			return nil, integration.ErrInvalidMsgType
		}
		msgServer := govkeeper.NewMsgServerImpl(f.govKeeper)
		resp, err := msgServer.ExecLegacyContent(ctx, msg)
		return resp, err
	}

	router.RegisterHandler(govSubmitProposalHandler, "/cosmos.gov.v1.MsgExecLegacyContent")
}

func (f *fixture) registerQueryRouterService(_ *msgrouter.RouterService) {
}
