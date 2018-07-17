package simulation

import (
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	gaia "github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/x/mock"
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

	// Run randomized simulation
	mock.RandomizedTesting(
		t, app.BaseApp,
		[]mock.TestAndRunTx{},
		[]mock.RandSetup{},
		[]mock.Invariant{},
		NumKeys,
		NumBlocks,
		BlockSize,
	)

}
