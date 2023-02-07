package simapp

import (
	"fmt"
	"os"
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
)

// Profile with:
// /usr/local/go/bin/go test -benchmem -run=^$ cosmossdk.io/simapp -bench ^BenchmarkFullAppSimulation$ -Commit=true -cpuprofile cpu.out
func BenchmarkFullAppSimulation(b *testing.B) {
	b.ReportAllocs()

	config := simcli.NewConfigFromFlags()
	config.ChainID = SimAppChainID

	db, dir, logger, skip, err := simtestutil.SetupSimulation(config, "goleveldb-app-sim", "Simulation", simcli.FlagVerboseValue, simcli.FlagEnabledValue)
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

	app := NewSimApp(logger, db, nil, true, appOptions, interBlockCacheOpt())

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		b,
		os.Stdout,
		app.BaseApp,
		AppStateFn(app.AppCodec(), app.SimulationManager()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		simtestutil.SimulationOperations(app, app.AppCodec(), config),
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

func BenchmarkInvariants(b *testing.B) {
	b.ReportAllocs()

	config := simcli.NewConfigFromFlags()
	config.ChainID = SimAppChainID

	db, dir, logger, skip, err := simtestutil.SetupSimulation(config, "leveldb-app-invariant-bench", "Simulation", simcli.FlagVerboseValue, simcli.FlagEnabledValue)
	if err != nil {
		b.Fatalf("simulation setup failed: %s", err.Error())
	}

	if skip {
		b.Skip("skipping benchmark application simulation")
	}

	config.AllInvariants = false

	defer func() {
		require.NoError(b, db.Close())
		require.NoError(b, os.RemoveAll(dir))
	}()

	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = DefaultNodeHome
	appOptions[server.FlagInvCheckPeriod] = simcli.FlagPeriodValue

	app := NewSimApp(logger, db, nil, true, appOptions, interBlockCacheOpt())

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		b,
		os.Stdout,
		app.BaseApp,
		AppStateFn(app.AppCodec(), app.SimulationManager()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		simtestutil.SimulationOperations(app, app.AppCodec(), config),
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

	ctx := app.NewContext(true, cmtproto.Header{Height: app.LastBlockHeight() + 1})

	// 3. Benchmark each invariant separately
	//
	// NOTE: We use the crisis keeper as it has all the invariants registered with
	// their respective metadata which makes it useful for testing/benchmarking.
	for _, cr := range app.CrisisKeeper.Routes() {
		cr := cr
		b.Run(fmt.Sprintf("%s/%s", cr.ModuleName, cr.Route), func(b *testing.B) {
			if res, stop := cr.Invar(ctx); stop {
				b.Fatalf(
					"broken invariant at block %d of %d\n%s",
					ctx.BlockHeight()-1, config.NumBlocks, res,
				)
			}
		})
	}
}
