package integration_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	mintkeeper "cosmossdk.io/x/mint/keeper"
	minttypes "cosmossdk.io/x/mint/types"

	"github.com/cosmos/cosmos-sdk/tests/integration/v2"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// Example shows how to use the integration test framework to test the integration of SDK modules.
// Panics are used in this example, but in a real test case, you should use the testing.T object and assertions.
// nolint:govet // ignore removal of parameter here as its run as a test as well.
func Example(t *testing.T) {
	t.Helper()
	authority := authtypes.NewModuleAddress("gov").String()

	var mintKeeper *mintkeeper.Keeper

	moduleConfigs := []configurator.ModuleOption{
		configurator.AccountsModule(),
		configurator.AuthModule(),
		configurator.BankModule(),
		configurator.VestingModule(),
		configurator.TxModule(),
		configurator.ValidateModule(),
		configurator.ConsensusModule(),
		configurator.GenutilModule(),
	}

	var err error
	startupCfg := integration.DefaultStartUpConfig(t)

	app, err := integration.NewApp(
		depinject.Configs(configurator.NewAppV2Config(moduleConfigs...), depinject.Provide(), depinject.Supply(log.NewNopLogger())),
		startupCfg,
		mintKeeper)
	require.NoError(t, err)
	require.NotNil(t, mintKeeper)

	ctx := app.StateLatestContext(t)

	mintMsgServer := mintkeeper.NewMsgServerImpl(mintKeeper)

	params := minttypes.DefaultParams()
	params.BlocksPerYear = 10000

	// now we can use the application to test a mint message
	result, err := app.RunMsg(t, ctx, func(ctx context.Context) (transaction.Msg, error) {
		msg := &minttypes.MsgUpdateParams{
			Authority: authority,
			Params:    params,
		}

		return mintMsgServer.UpdateParams(ctx, msg)
	})
	if err != nil {
		panic(err)
	}

	// in this example the result is an empty response, a nil check is enough
	// in other cases, it is recommended to check the result value.
	if result == nil {
		panic(errors.New("unexpected nil result"))
	}

	_, ok := result.(*minttypes.MsgUpdateParamsResponse)
	require.True(t, ok)

	// we should also check the state of the application
	got, err := mintKeeper.Params.Get(ctx)
	if err != nil {
		panic(err)
	}

	if diff := cmp.Diff(got, params); diff != "" {
		panic(diff)
	}
	fmt.Println(got.BlocksPerYear)
	// Output: 10000
}
