//go:build sims

package simapp

import (
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/testutils/sims"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

// Profile with:
// /usr/local/go/bin/go test -benchmem -run=^$ cosmossdk.io/simapp -bench ^BenchmarkFullAppSimulation$ -Commit=true -cpuprofile cpu.out
func BenchmarkFullAppSimulation(b *testing.B) {
	b.ReportAllocs()

	config := simcli.NewConfigFromFlags()
	config.ChainID = sims.SimAppChainID

	db, dir, logger, skip, err := simtestutil.SetupSimulation(config, "goleveldb-app-sim", "Simulation", simcli.FlagVerboseValue, true)
	if err != nil {
		b.Fatalf("simulation setup failed: %s", err.Error())
	}

	if skip {
		b.Skip("skipping benchmark application simulation")
	}

	defer func() {
		require.NoError(b, db.Close())
		require.NoError(b, os.RemoveAll(dir))
	}()

	appOptions := viper.New()
	appOptions.SetDefault(flags.FlagHome, DefaultNodeHome)
	appOptions.SetDefault(server.FlagInvCheckPeriod, simcli.FlagPeriodValue)

	app := NewSimApp(logger, db, nil, true, appOptions, interBlockCacheOpt(), baseapp.SetChainID(sims.SimAppChainID))

	// run randomized simulation
	simParams, simErr := simulation.SimulateFromSeedX(
		b,
		log.NewNopLogger(),
		os.Stdout,
		app.BaseApp,
		simtestutil.AppStateFn(app.AppCodec(), app.AccountKeeper.AddressCodec(), app.StakingKeeper.ValidatorAddressCodec(), app.SimulationManager(), app.DefaultGenesis()),
		simtypes.RandomAccounts,
		simtestutil.SimulationOperations(app, app.AppCodec(), config, app.txConfig),
		BlockedAddresses(),
		config,
		app.AppCodec(),
		&simulation.DummyLogWriter{},
	)

	// export state and simParams before the simulation error is checked
	if err = simtestutil.CheckExportSimulation(app, config, simParams); err != nil {
		b.Fatal(err)
	}

	if simErr != nil {
		b.Fatal(simErr)
	}

	if config.Commit {
		simtestutil.PrintStats(db)
	}
}

func BenchmarkInvariants(b *testing.B) {
	b.ReportAllocs()

	config := simcli.NewConfigFromFlags()
	config.ChainID = sims.SimAppChainID

	db, dir, logger, skip, err := simtestutil.SetupSimulation(config, "leveldb-app-invariant-bench", "Simulation", simcli.FlagVerboseValue, true)
	if err != nil {
		b.Fatalf("simulation setup failed: %s", err.Error())
	}

	if skip {
		b.Skip("skipping benchmark application simulation")
	}

	defer func() {
		require.NoError(b, db.Close())
		require.NoError(b, os.RemoveAll(dir))
	}()

	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = DefaultNodeHome
	appOptions[server.FlagInvCheckPeriod] = simcli.FlagPeriodValue

	app := NewSimApp(logger, db, nil, true, appOptions, interBlockCacheOpt(), baseapp.SetChainID(sims.SimAppChainID))

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		b,
		os.Stdout,
		app.BaseApp,
		simtestutil.AppStateFn(app.AppCodec(), app.AccountKeeper.AddressCodec(), app.StakingKeeper.ValidatorAddressCodec(), app.SimulationManager(), app.DefaultGenesis()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		simtestutil.SimulationOperations(app, app.AppCodec(), config, app.txConfig),
		BlockedAddresses(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	if err = simtestutil.CheckExportSimulation(app, config, simParams); err != nil {
		b.Fatal(err)
	}

	if simErr != nil {
		b.Fatal(simErr)
	}

	if config.Commit {
		simtestutil.PrintStats(db)
	}
}
