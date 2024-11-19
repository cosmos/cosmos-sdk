package auth

import (
	"context"
	"testing"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/x/accounts"
	_ "cosmossdk.io/x/accounts" // import as blank for app wiring
	basedepinject "cosmossdk.io/x/accounts/defaults/base/depinject"
	_ "cosmossdk.io/x/bank" // import as blank for app wiring
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	_ "cosmossdk.io/x/consensus" // import as blank for app wiring
	_ "cosmossdk.io/x/staking"   // import as blank for app wirings

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	_ "github.com/cosmos/cosmos-sdk/x/auth" // import as blank for app wiring
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import as blank for app wiring``
	_ "github.com/cosmos/cosmos-sdk/x/auth/vesting"   // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/genutil"        // import as blank for app wiring
	"github.com/stretchr/testify/require"
)

type suite struct {
	app *integration.App

	cdc codec.Codec
	ctx context.Context

	authKeeper     authkeeper.AccountKeeper
	accountsKeeper accounts.Keeper
	bankKeeper     bankkeeper.Keeper

	routerService *integration.RouterService
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

	res.routerService = integration.NewRouterService()
	res.registerRouterService()

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
		), depinject.Supply(log.NewNopLogger(), &integration.BranchService{}, res.routerService)),
		startupCfg,
		&res.bankKeeper, &res.accountsKeeper, &res.authKeeper)
	require.NoError(t, err)

	res.ctx = res.app.StateLatestContext(t)

	return &res
}

func (s *suite) registerRouterService() {
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

	s.routerService.RegisterHandler(bankSendHandler, "cosmos.bank.v1beta1.MsgSend")
}
