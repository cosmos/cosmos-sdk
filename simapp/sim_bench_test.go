package simapp

import (
	"fmt"
	"os"
	"testing"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
)

// Profile with:
// /usr/local/go/bin/go test -benchmem -run=^$ github.com/cosmos/cosmos-sdk/simapp -bench ^BenchmarkFullAppSimulation$ -Commit=true -cpuprofile cpu.out
func BenchmarkFullAppSimulation(b *testing.B) {
	config, db, dir, _, _, err := SetupSimulation("goleveldb-app-sim", "Simulation")
	if err != nil {
		fmt.Println(err, "simulation setup failed")
		b.Fail()
	}

	defer func() {
		db.Close()
		err = os.RemoveAll(dir)
		if err != nil {
			fmt.Println(err.Error())
			b.Fail()
		}
	}()

	logger := log.NewNopLogger()
	app := NewSimApp(logger, db, nil, true, FlagPeriodValue, interBlockCacheOpt())

	err = RunSimulation(b, os.Stdout, db, app, app.BaseApp, config)
	if err != nil {
		fmt.Println(err)
		b.FailNow()
	}
}

func BenchmarkInvariants(b *testing.B) {
	config, db, dir, _, _, err := SetupSimulation("leveldb-app-invariant-bench", "Simulation")
	if err != nil {
		fmt.Println(err, "simulation setup failed")
		b.Fail()
	}

	logger := log.NewNopLogger()
	config.AllInvariants = false

	defer func() {
		db.Close()
		err = os.RemoveAll(dir)
		if err != nil {
			fmt.Println(err.Error())
			b.Fail()
		}
	}()

	app := NewSimApp(logger, db, nil, true, FlagPeriodValue, interBlockCacheOpt())

	// 2. Run parameterized simulation (w/o invariants)
	err = RunSimulation(b, os.Stdout, db, app, app.BaseApp, config)
	if err != nil {
		fmt.Println(err)
		b.FailNow()
	}

	ctx := app.NewContext(true, abci.Header{Height: app.LastBlockHeight() + 1})

	// 3. Benchmark each invariant separately
	//
	// NOTE: We use the crisis keeper as it has all the invariants registered with
	// their respective metadata which makes it useful for testing/benchmarking.
	for _, cr := range app.CrisisKeeper.Routes() {
		cr := cr
		b.Run(fmt.Sprintf("%s/%s", cr.ModuleName, cr.Route), func(b *testing.B) {
			if res, stop := cr.Invar(ctx); stop {
				fmt.Printf("broken invariant at block %d of %d\n%s", ctx.BlockHeight()-1, config.NumBlocks, res)
				b.FailNow()
			}
		})
	}
}
