package app

import (
	"encoding/json"
	"flag"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banksim "github.com/cosmos/cosmos-sdk/x/bank/simulation"
	govsim "github.com/cosmos/cosmos-sdk/x/gov/simulation"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	slashingsim "github.com/cosmos/cosmos-sdk/x/slashing/simulation"
	stake "github.com/cosmos/cosmos-sdk/x/stake"
	stakesim "github.com/cosmos/cosmos-sdk/x/stake/simulation"
)

var (
	seed      int64
	numBlocks int
	blockSize int
	enabled   bool
	verbose   bool
)

func init() {
	flag.Int64Var(&seed, "SimulationSeed", 42, "Simulation random seed")
	flag.IntVar(&numBlocks, "SimulationNumBlocks", 500, "Number of blocks")
	flag.IntVar(&blockSize, "SimulationBlockSize", 200, "Operations per block")
	flag.BoolVar(&enabled, "SimulationEnabled", false, "Enable the simulation")
	flag.BoolVar(&verbose, "SimulationVerbose", false, "Verbose log output")
}

func appStateFn(r *rand.Rand, keys []crypto.PrivKey, accs []sdk.AccAddress) json.RawMessage {
	var genesisAccounts []GenesisAccount

	// Randomly generate some genesis accounts
	for _, acc := range accs {
		coins := sdk.Coins{sdk.Coin{"steak", sdk.NewInt(100)}}
		genesisAccounts = append(genesisAccounts, GenesisAccount{
			Address: acc,
			Coins:   coins,
		})
	}

	// Default genesis state
	stakeGenesis := stake.DefaultGenesisState()
	var validators []stake.Validator
	var delegations []stake.Delegation
	// XXX Try different numbers of initially bonded validators
	numInitiallyBonded := int64(50)
	for i := 0; i < int(numInitiallyBonded); i++ {
		validator := stake.NewValidator(accs[i], keys[i].PubKey(), stake.Description{})
		validator.Tokens = sdk.NewRat(100)
		validator.DelegatorShares = sdk.NewRat(100)
		delegation := stake.Delegation{accs[i], accs[i], sdk.NewRat(100), 0}
		validators = append(validators, validator)
		delegations = append(delegations, delegation)
	}
	stakeGenesis.Pool.LooseTokens = sdk.NewRat(int64(100*250) + (numInitiallyBonded * 100))
	stakeGenesis.Validators = validators
	stakeGenesis.Bonds = delegations
	// No inflation, for now
	stakeGenesis.Params.InflationMax = sdk.NewRat(0)
	stakeGenesis.Params.InflationMin = sdk.NewRat(0)
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
	var logger log.Logger
	if verbose {
		logger = log.TestingLogger()
	} else {
		logger = log.NewNopLogger()
	}
	db := dbm.NewMemDB()
	app := NewGaiaApp(logger, db, nil)
	require.Equal(t, "GaiaApp", app.Name())

	allInvariants := func(t *testing.T, baseapp *baseapp.BaseApp, log string) {
		banksim.NonnegativeBalanceInvariant(app.accountMapper)(t, baseapp, log)
		govsim.AllInvariants()(t, baseapp, log)
		stakesim.AllInvariants(app.coinKeeper, app.stakeKeeper, app.accountMapper)(t, baseapp, log)
		slashingsim.AllInvariants()(t, baseapp, log)
	}

	// Run randomized simulation
	simulation.SimulateFromSeed(
		t, app.BaseApp, appStateFn, seed,
		[]simulation.TestAndRunTx{
			banksim.TestAndRunSingleInputMsgSend(app.accountMapper),
			govsim.SimulateMsgSubmitProposal(app.govKeeper, app.stakeKeeper),
			govsim.SimulateMsgDeposit(app.govKeeper, app.stakeKeeper),
			govsim.SimulateMsgVote(app.govKeeper, app.stakeKeeper),
			stakesim.SimulateMsgCreateValidator(app.accountMapper, app.stakeKeeper),
			stakesim.SimulateMsgEditValidator(app.stakeKeeper),
			stakesim.SimulateMsgDelegate(app.accountMapper, app.stakeKeeper),
			stakesim.SimulateMsgBeginUnbonding(app.accountMapper, app.stakeKeeper),
			stakesim.SimulateMsgCompleteUnbonding(app.stakeKeeper),
			stakesim.SimulateMsgBeginRedelegate(app.accountMapper, app.stakeKeeper),
			stakesim.SimulateMsgCompleteRedelegate(app.stakeKeeper),
			slashingsim.SimulateMsgUnrevoke(app.slashingKeeper),
		},
		[]simulation.RandSetup{},
		[]simulation.Invariant{
			allInvariants,
		},
		numBlocks,
		blockSize,
	)

}

// TODO: Make this not depend on Gaia or any of the modules,
// and place it in random_simulation_test.go
//
// Test doesn't use `app.ExportAppStateAndValidators` as that panics with the following:
// panic: Stored pool should not have been nil [recovered]
//	panic: Stored pool should not have been nil
// Change to `app.ExportAppStateAndValidators` once it is fixed
func TestAppStateDeterminism(t *testing.T) {
	numTimesToRun := 5
	appHashList := make([]json.RawMessage, numTimesToRun)

	seed := rand.Int63()
	for i := 0; i < numTimesToRun; i++ {
		logger := log.NewNopLogger()
		db := dbm.NewMemDB()
		app := NewGaiaApp(logger, db, nil)

		noOpInvariant := func(t *testing.T, baseapp *baseapp.BaseApp, log string) {}
		noOpTestAndRunTx := func(t *testing.T, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
			privKeys []crypto.PrivKey, log string, event func(string),
		) (action string, err sdk.Error) {
			return "", nil
		}

		// Run randomized simulation
		simulation.SimulateFromSeed(
			t, app.BaseApp, appStateFn, seed,
			[]simulation.TestAndRunTx{
				noOpTestAndRunTx,
			},
			[]simulation.RandSetup{},
			[]simulation.Invariant{noOpInvariant},
			0,
			10,
		)
		appHash := app.LastCommitID().Hash
		appHashList[i] = appHash
	}
	for i := 1; i < numTimesToRun; i++ {
		require.Equal(t, appHashList[0], appHashList[i])
	}
}
