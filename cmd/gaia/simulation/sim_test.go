package simulation

import (
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	gaia "github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	stake "github.com/cosmos/cosmos-sdk/x/stake"
)

const (
	NumKeys   = 10
	NumBlocks = 1000
	BlockSize = 1000
)

func TestFullGaiaSimulation(t *testing.T) {

	// Setup Gaia application
	logger := log.NewNopLogger()
	db := dbm.NewMemDB()
	app := gaia.NewGaiaApp(logger, db, nil)
	require.Equal(t, "GaiaApp", app.Name())

	// Default genesis state
	genesis := gaia.GenesisState{
		Accounts:  []gaia.GenesisAccount{},
		StakeData: stake.DefaultGenesisState(),
	}

	// Marshal genesis
	appState, err := gaia.MakeCodec().MarshalJSON(genesis)
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
