package app

import (
	"encoding/json"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banksim "github.com/cosmos/cosmos-sdk/x/bank/simulation"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	stake "github.com/cosmos/cosmos-sdk/x/stake"
	stakesim "github.com/cosmos/cosmos-sdk/x/stake/simulation"
)

const (
	defaultNumKeys   = 10
	defaultNumBlocks = 100
	defaultBlockSize = 100

	simulationEnvEnable    = "GAIA_SIMULATION_ENABLED"
	simulationEnvSeed      = "GAIA_SIMULATION_SEED"
	simulationEnvKeys      = "GAIA_SIMULATION_KEYS"
	simulationEnvBlocks    = "GAIA_SIMULATION_BLOCKS"
	simulationEnvBlockSize = "GAIA_SIMULATION_BLOCK_SIZE"
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
	stakeGenesis := stake.DefaultGenesisState()
	stakeGenesis.Pool.LooseTokens = sdk.NewRat(1000)
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
	if os.Getenv(simulationEnvEnable) == "" {
		t.Skip("Skipping Gaia simulation")
	}

	// Setup Gaia application
	logger := log.NewNopLogger()
	db := dbm.NewMemDB()
	app := NewGaiaApp(logger, db, nil)
	require.Equal(t, "GaiaApp", app.Name())

	var seed int64
	var err error
	envSeed := os.Getenv(simulationEnvSeed)
	if envSeed != "" {
		seed, err = strconv.ParseInt(envSeed, 10, 64)
		require.Nil(t, err)
	} else {
		seed = time.Now().UnixNano()
	}

	keys := defaultNumKeys
	envKeys := os.Getenv(simulationEnvKeys)
	if envKeys != "" {
		keys, err = strconv.Atoi(envKeys)
		require.Nil(t, err)
	}

	blocks := defaultNumBlocks
	envBlocks := os.Getenv(simulationEnvBlocks)
	if envBlocks != "" {
		blocks, err = strconv.Atoi(envBlocks)
		require.Nil(t, err)
	}

	blockSize := defaultBlockSize
	envBlockSize := os.Getenv(simulationEnvBlockSize)
	if envBlockSize != "" {
		blockSize, err = strconv.Atoi(envBlockSize)
		require.Nil(t, err)
	}

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
		keys,
		blocks,
		blockSize,
	)

}
