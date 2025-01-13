//go:build sims

package simapp

import (
	"bytes"
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	genutil "github.com/cosmos/cosmos-sdk/x/genutil/v2"
	simcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
)

func init() {
	simcli.GetSimulatorFlags()
}

func TestFullAppSimulation(t *testing.T) {
	RunWithSeeds[Tx](t, NewSimApp[Tx], AppConfig, DefaultSeeds)
}

func TestAppStateDeterminism(t *testing.T) {
	var seeds []int64
	if s := simcli.NewConfigFromFlags().Seed; s != simcli.DefaultSeedValue {
		// override defaults with user data
		seeds = []int64{s, s, s} // run same simulation 3 times
	} else {
		seeds = []int64{ // some random seeds, tripled to ensure same app-hash on all runs
			1, 1, 1,
			3, 3, 3,
			5, 5, 5,
		}
	}

	var mx sync.Mutex
	appHashResults := make(map[int64][]byte)
	captureAndCheckHash := func(tb testing.TB, cs ChainState[Tx], ti TestInstance[Tx], _ []simtypes.Account) {
		tb.Helper()
		mx.Lock()
		defer mx.Unlock()
		otherHashes, ok := appHashResults[ti.Seed]
		if !ok {
			appHashResults[ti.Seed] = cs.AppHash
			return
		}
		if !bytes.Equal(otherHashes, cs.AppHash) {
			tb.Fatalf("non-determinism in seed %d", ti.Seed)
		}
	}
	// run simulations
	RunWithSeeds(t, NewSimApp[Tx], AppConfig, seeds, captureAndCheckHash)
}

// ExportableApp defines an interface for exporting application state and validator set.
type ExportableApp interface {
	ExportAppStateAndValidators(forZeroHeight bool, jailAllowedAddrs []string) (genutil.ExportedApp, error)
}

// Scenario:
//
//	Start a fresh node and run n blocks, export state
//	set up a new node instance, Init chain from exported genesis
//	run new instance for n blocks
func TestAppSimulationAfterImport(t *testing.T) {
	appFactory := NewSimApp[Tx]
	cfg := simcli.NewConfigFromFlags()
	cfg.ChainID = SimAppChainID

	exportAndStartChainFromGenesisPostAction := func(tb testing.TB, cs ChainState[Tx], ti TestInstance[Tx], accs []simtypes.Account) {
		tb.Helper()
		tb.Log("exporting genesis...\n")
		app, ok := ti.App.(ExportableApp)
		require.True(tb, ok)
		exported, err := app.ExportAppStateAndValidators(false, []string{})
		require.NoError(tb, err)

		genesisTimestamp := cs.BlockTime.Add(24 * time.Hour)
		startHeight := uint64(exported.Height + 1)
		chainID := SimAppChainID + "_2"

		importGenesisChainStateFactory := func(ctx context.Context, r *rand.Rand) (TestInstance[Tx], ChainState[Tx], []simtypes.Account) {
			testInstance := SetupTestInstance(tb, appFactory, AppConfig, ti.Seed)
			newCs := testInstance.InitializeChain(
				tb,
				ctx,
				chainID,
				genesisTimestamp,
				startHeight,
				exported.AppState,
			)
			return testInstance, newCs, accs
		}
		// run sims with new app setup from exported genesis
		RunWithSeedX[Tx](tb, cfg, importGenesisChainStateFactory, ti.Seed)
	}
	RunWithSeeds[Tx, *SimApp[Tx]](t, appFactory, AppConfig, DefaultSeeds, exportAndStartChainFromGenesisPostAction)
}
