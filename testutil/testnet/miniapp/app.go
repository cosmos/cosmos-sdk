// Package miniapp provides a bare minimum app intended for exercising the testnet framework.
//
// This is a non-internal package so that example tests can be copied, pasted, and run
// without further configuration, to see a testnet running immediately.
// Otherwise, tests that utilize the testnet framework
// should provide a more realistic app with behavior that needs to be tested.
package miniapp

import (
	"fmt"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"

	// Required for side effects of module registration, for configurator.
	_ "github.com/cosmos/cosmos-sdk/x/auth"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	_ "github.com/cosmos/cosmos-sdk/x/bank"
	_ "github.com/cosmos/cosmos-sdk/x/consensus"
	_ "github.com/cosmos/cosmos-sdk/x/genutil"
	_ "github.com/cosmos/cosmos-sdk/x/params"
	_ "github.com/cosmos/cosmos-sdk/x/staking"
)

var appConfig = configurator.NewAppConfig(
	configurator.AuthModule(),
	configurator.ParamsModule(),
	configurator.BankModule(),
	configurator.GenutilModule(),
	configurator.StakingModule(),
	configurator.ConsensusModule(),
	configurator.TxModule(),
)

// New returns a new minimal app for exercising the testnet framework.
func New(
	logger log.Logger,
	db dbm.DB,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *runtime.App {
	var builder *runtime.AppBuilder

	if err := depinject.Inject(appConfig, &builder); err != nil {
		panic(fmt.Errorf("failed to inject dependencies: %w", err))
	}

	app := builder.Build(logger, db, nil, baseAppOptions...)

	if err := app.Load(true); err != nil {
		panic(fmt.Errorf("failed to load app: %w", err))
	}

	return app
}
