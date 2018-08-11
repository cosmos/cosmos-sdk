package app

import (
	"encoding/json"
	"flag"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banksim "github.com/cosmos/cosmos-sdk/x/bank/simulation"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	stake "github.com/cosmos/cosmos-sdk/x/stake"
	stakesim "github.com/cosmos/cosmos-sdk/x/stake/simulation"
)

var (
	seed      int64
	numKeys   int
	numBlocks int
	blockSize int
	enabled   bool
)

func init() {
	flag.Int64Var(&seed, "SimulationSeed", 42, "Simulation random seed")
	flag.IntVar(&numKeys, "SimulationNumKeys", 10, "Number of keys (accounts)")
	flag.IntVar(&numBlocks, "SimulationNumBlocks", 100, "Number of blocks")
	flag.IntVar(&blockSize, "SimulationBlockSize", 100, "Operations per block")
	flag.BoolVar(&enabled, "SimulationEnabled", false, "Enable the simulation")
}

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
	stakeGenesis := stake.DefaultGenesisState()
	stakeGenesis.Pool.LooseTokens = sdk.NewDec(1000)
	genesis := GenesisState{
		Accounts:  genesisAccounts,
		StakeData: stakeGenesis,
	}

	// Marshal genesis
	appState, err := MakeCodec().MarshalJSON(genesis)
	if err != nil {
		panic(err)
	}

	return appState
}

func TestFullGaiaSimulation(t *testing.T) {
	if !enabled {
		t.Skip("Skipping Gaia simulation")
	}

	// Setup Gaia application
	logger := log.NewNopLogger()
	db := dbm.NewMemDB()
	app := NewGaiaApp(logger, db, nil)
	require.Equal(t, "GaiaApp", app.Name())

	// Run randomized simulation
	simulation.SimulateFromSeed(
		t, app.BaseApp, appStateFn, seed,
		[]simulation.TestAndRunTx{
			banksim.TestAndRunSingleInputMsgSend(app.accountMapper),
			stakesim.SimulateMsgCreateValidator(app.accountMapper, app.stakeKeeper),
			stakesim.SimulateMsgEditValidator(app.stakeKeeper),
			stakesim.SimulateMsgDelegate(app.accountMapper, app.stakeKeeper),
			stakesim.SimulateMsgBeginUnbonding(app.accountMapper, app.stakeKeeper),
			stakesim.SimulateMsgCompleteUnbonding(app.stakeKeeper),
			stakesim.SimulateMsgBeginRedelegate(app.accountMapper, app.stakeKeeper),
			stakesim.SimulateMsgCompleteRedelegate(app.stakeKeeper),
		},
		[]simulation.RandSetup{},
		[]simulation.Invariant{
			banksim.NonnegativeBalanceInvariant(app.accountMapper),
			stakesim.AllInvariants(app.coinKeeper, app.stakeKeeper, app.accountMapper),
		},
		numKeys,
		numBlocks,
		blockSize,
	)

}
