package simapp

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Profile with:
// /usr/local/go/bin/go test -benchmem -run=^$ github.com/cosmos/cosmos-sdk/simapp -bench ^BenchmarkFullAppSimulation$ -Commit=true -cpuprofile cpu.out
func BenchmarkFullAppSimulation(b *testing.B) {
	logger := log.NewNopLogger()
	config := NewConfigFromFlags()

	var db dbm.DB
	dir, _ := ioutil.TempDir("", "goleveldb-app-sim")
	db, _ = sdk.NewLevelDB("Simulation", dir)
	defer func() {
		db.Close()
		os.RemoveAll(dir)
	}()

	app := NewSimApp(logger, db, nil, true, FlagPeriodValue, interBlockCacheOpt())

	// Run randomized simulation
	// TODO: parameterize numbers, save for a later PR
	_, simParams, simErr := simulation.SimulateFromSeed(
		b, os.Stdout, app.BaseApp, AppStateFn(app.Codec(), app.sm),
		testAndRunTxs(app, config), app.ModuleAccountAddrs(), config,
	)

	// export state and params before the simulation error is checked
	if config.ExportStatePath != "" {
		if err := ExportStateToJSON(app, config.ExportStatePath); err != nil {
			fmt.Println(err)
			b.Fail()
		}
	}

	if config.ExportParamsPath != "" {
		if err := ExportParamsToJSON(simParams, config.ExportParamsPath); err != nil {
			fmt.Println(err)
			b.Fail()
		}
	}

	if simErr != nil {
		fmt.Println(simErr)
		b.FailNow()
	}

	if config.Commit {
		fmt.Println("\nGoLevelDB Stats")
		fmt.Println(db.Stats()["leveldb.stats"])
		fmt.Println("GoLevelDB cached block size", db.Stats()["leveldb.cachedblock"])
	}
}

func BenchmarkInvariants(b *testing.B) {
	logger := log.NewNopLogger()

	config := NewConfigFromFlags()
	config.AllInvariants = false

	dir, _ := ioutil.TempDir("", "goleveldb-app-invariant-bench")
	db, _ := sdk.NewLevelDB("simulation", dir)

	defer func() {
		db.Close()
		os.RemoveAll(dir)
	}()

	app := NewSimApp(logger, db, nil, true, FlagPeriodValue, interBlockCacheOpt())

	// 2. Run parameterized simulation (w/o invariants)
	_, simParams, simErr := simulation.SimulateFromSeed(
		b, ioutil.Discard, app.BaseApp, AppStateFn(app.Codec(), app.sm),
		testAndRunTxs(app, config), app.ModuleAccountAddrs(), config,
	)

	// export state and params before the simulation error is checked
	if config.ExportStatePath != "" {
		if err := ExportStateToJSON(app, config.ExportStatePath); err != nil {
			fmt.Println(err)
			b.Fail()
		}
	}

	if config.ExportParamsPath != "" {
		if err := ExportParamsToJSON(simParams, config.ExportParamsPath); err != nil {
			fmt.Println(err)
			b.Fail()
		}
	}

	if simErr != nil {
		fmt.Println(simErr)
		b.FailNow()
	}

	ctx := app.NewContext(true, abci.Header{Height: app.LastBlockHeight() + 1})

	// 3. Benchmark each invariant separately
	//
	// NOTE: We use the crisis keeper as it has all the invariants registered with
	// their respective metadata which makes it useful for testing/benchmarking.
	for _, cr := range app.CrisisKeeper.Routes() {
		b.Run(fmt.Sprintf("%s/%s", cr.ModuleName, cr.Route), func(b *testing.B) {
			if res, stop := cr.Invar(ctx); stop {
				fmt.Printf("broken invariant at block %d of %d\n%s", ctx.BlockHeight()-1, config.NumBlocks, res)
				b.FailNow()
			}
		})
	}
}
