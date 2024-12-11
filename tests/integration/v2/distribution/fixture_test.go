package distribution

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/comet"
	corecontext "cosmossdk.io/core/context"
	"cosmossdk.io/core/router"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/runtime/v2"
	_ "cosmossdk.io/x/accounts" // import as blank for app wiring
	_ "cosmossdk.io/x/bank"     // import as blank for app wiring
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	_ "cosmossdk.io/x/consensus"    // import as blank for app wiring
	_ "cosmossdk.io/x/distribution" // import as blank for app wiring
	distrkeeper "cosmossdk.io/x/distribution/keeper"
	_ "cosmossdk.io/x/mint"         // import as blank for app wiring
	_ "cosmossdk.io/x/protocolpool" // import as blank for app wiring
	poolkeeper "cosmossdk.io/x/protocolpool/keeper"
	_ "cosmossdk.io/x/staking" // import as blank for app wiring
	stakingkeeper "cosmossdk.io/x/staking/keeper"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/x/auth" // import as blank for app wiring
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import as blank for app wiring``
	_ "github.com/cosmos/cosmos-sdk/x/auth/vesting"   // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/genutil"        // import as blank for app wiring
)

var (
	emptyDelAddr sdk.AccAddress
	emptyValAddr sdk.ValAddress
)

var (
	PKS = simtestutil.CreateTestPubKeys(3)

	valConsPk0 = PKS[0]
)

type fixture struct {
	app *integration.App

	ctx context.Context
	cdc codec.Codec

	queryClient distrkeeper.Querier

	authKeeper    authkeeper.AccountKeeper
	bankKeeper    bankkeeper.Keeper
	distrKeeper   distrkeeper.Keeper
	stakingKeeper *stakingkeeper.Keeper
	poolKeeper    poolkeeper.Keeper

	addr    sdk.AccAddress
	valAddr sdk.ValAddress
}

func createTestFixture(t *testing.T) *fixture {
	t.Helper()
	res := fixture{}

	moduleConfigs := []configurator.ModuleOption{
		configurator.AccountsModule(),
		configurator.AuthModule(),
		configurator.BankModule(),
		configurator.StakingModule(),
		configurator.TxModule(),
		configurator.ValidateModule(),
		configurator.ConsensusModule(),
		configurator.GenutilModule(),
		configurator.DistributionModule(),
		configurator.MintModule(),
		configurator.ProtocolPoolModule(),
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
	startupCfg.HeaderService = &integration.HeaderService{}

	res.app, err = integration.NewApp(
		depinject.Configs(configurator.NewAppV2Config(moduleConfigs...), depinject.Supply(log.NewNopLogger())),
		startupCfg,
		&res.bankKeeper, &res.distrKeeper, &res.authKeeper, &res.stakingKeeper, &res.poolKeeper, &res.cdc)
	require.NoError(t, err)

	addr := sdk.AccAddress(PKS[0].Address())
	valAddr := sdk.ValAddress(addr)
	valConsAddr := sdk.ConsAddress(valConsPk0.Address())

	ctx := res.app.StateLatestContext(t)
	res.addr = addr
	res.valAddr = valAddr

	// set proposer and vote infos
	res.ctx = context.WithValue(ctx, corecontext.CometInfoKey, comet.Info{
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
		ProposerAddress: valConsAddr,
	})

	res.queryClient = distrkeeper.NewQuerier(res.distrKeeper)

	return &res
}

func (s *fixture) registerMsgRouterService(router *integration.RouterService) {
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

func (s *fixture) registerQueryRouterService(router *integration.RouterService) {
	// register custom router service
}
