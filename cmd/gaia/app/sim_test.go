package app

import (
	"encoding/json"
	"math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	stake "github.com/cosmos/cosmos-sdk/x/stake"
	stakesim "github.com/cosmos/cosmos-sdk/x/stake/simulation"
)

const (
	NumKeys   = 10
	NumBlocks = 1000
	BlockSize = 1000

	simulationEnv = "ENABLE_GAIA_SIMULATION"
)

func appStateFn(r *rand.Rand, accs []sdk.AccAddress) json.RawMessage {
	var genesisAccounts []GenesisAccount

	// Randomly generate some genesis accounts
	for _, addr := range accs {
		coins := sdk.Coins{sdk.Coin{"steak", sdk.NewInt(100)}}
		genesisAccounts = append(genesisAccounts, GenesisAccount{
			Address: addr,
			Coins:   coins,
		})
	}

	// Default genesis state
	genesis := GenesisState{
		Accounts:  genesisAccounts,
		StakeData: stake.DefaultGenesisState(),
	}

	// Marshal genesis
	appState, err := MakeCodec().MarshalJSON(genesis)
	if err != nil {
		panic(err)
	}

	return appState
}

func TestFullGaiaSimulation(t *testing.T) {
	if os.Getenv(simulationEnv) == "" {
		t.Skip("Skipping Gaia simulation")
	}

	// Setup Gaia application
	logger := log.NewNopLogger()
	db := dbm.NewMemDB()
	app := NewGaiaApp(logger, db, nil)
	require.Equal(t, "GaiaApp", app.Name())

	// Run randomized simulation
	simulation.Simulate(
		t, app.BaseApp, appStateFn,
		[]simulation.TestAndRunTx{
			stakesim.SimulateMsgCreateValidator(app.accountMapper, app.stakeKeeper),
		},
		[]simulation.RandSetup{},
		[]simulation.Invariant{
			stakesim.AllInvariants(app.coinKeeper, app.stakeKeeper, app.accountMapper),
		},
		NumKeys,
		NumBlocks,
		BlockSize,
	)

}
