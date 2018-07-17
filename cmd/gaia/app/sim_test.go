package app

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	stake "github.com/cosmos/cosmos-sdk/x/stake"
)

const (
	NumKeys   = 10
	NumBlocks = 1000
	BlockSize = 1000

	simulationEnv = "ENABLE_GAIA_SIMULATION"
)

func TestFullGaiaSimulation(t *testing.T) {
	if os.Getenv(simulationEnv) == "" {
		t.Skip("Skipping Gaia simulation")
	}

	// Setup Gaia application
	logger := log.NewNopLogger()
	db := dbm.NewMemDB()
	app := NewGaiaApp(logger, db, nil)
	require.Equal(t, "GaiaApp", app.Name())

	// Default genesis state
	genesis := GenesisState{
		Accounts:  []GenesisAccount{},
		StakeData: stake.DefaultGenesisState(),
	}

	// Marshal genesis
	appState, err := MakeCodec().MarshalJSON(genesis)
	if err != nil {
		panic(err)
	}

	// Run randomized simulation
	simulation.Simulate(
		t, app.BaseApp, appState,
		[]simulation.TestAndRunTx{},
		[]simulation.RandSetup{},
		[]simulation.Invariant{},
		NumKeys,
		NumBlocks,
		BlockSize,
	)

}
