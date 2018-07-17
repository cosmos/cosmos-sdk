package simulation

import (
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	gaia "github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	// stakesim "github.com/cosmos/cosmos-sdk/x/stake/simulation"
)

func TestFullGaiaSimulation(t *testing.T) {
	// Setup Gaia application
	logger := log.NewNopLogger()
	db := dbm.NewMemDB()
	app := gaia.NewGaiaApp(logger, db, nil)
	require.Equal(t, "GaiaApp", app.Name())
}
