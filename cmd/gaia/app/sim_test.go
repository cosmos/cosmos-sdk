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
	seed             int64
	numKeys          int
	numBlocks        int
	blockSize        int
	minTimePerBlock  int64
	maxTimePerBlock  int64
	signingFraction  float64
	evidenceFraction float64
	enabled          bool
)

func init() {
	flag.Int64Var(&seed, "SimulationSeed", 42, "Simulation random seed")
	flag.IntVar(&numKeys, "SimulationNumKeys", 500, "Number of keys (accounts)")
	flag.IntVar(&numBlocks, "SimulationNumBlocks", 500, "Number of blocks")
	flag.IntVar(&blockSize, "SimulationBlockSize", 200, "Operations per block")
	flag.Int64Var(&minTimePerBlock, "SimulationMinTimePerBlock", 86400, "Minimum time per block (seconds)")
	flag.Int64Var(&maxTimePerBlock, "SimulationMaxTimePerBlock", 2*86400, "Maximum time per block (seconds)")
	flag.Float64Var(&signingFraction, "SimulationSigningFraction", 0.7, "Chance a given validator signs a given block")
	flag.Float64Var(&evidenceFraction, "SimulationEvidenceFraction", 0.01, "Chance that any evidence is found on a given block")
	flag.BoolVar(&enabled, "SimulationEnabled", false, "Enable the simulation")
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
	numInitiallyBonded := int64(50)
	for i := 0; i < int(numInitiallyBonded); i++ {
		validator := stake.NewValidator(accs[i], keys[i].PubKey(), stake.Description{})
		validator.Tokens = sdk.NewRat(100)
		validator.DelegatorShares = sdk.NewRat(100)
		delegation := stake.Delegation{accs[i], accs[i], sdk.NewRat(100), 0}
		validators = append(validators, validator)
		delegations = append(delegations, delegation)
	}
	stakeGenesis.Pool.LooseTokens = sdk.NewRat(int64(100*numKeys) + (numInitiallyBonded * 100))
	stakeGenesis.Validators = validators
	stakeGenesis.Bonds = delegations
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
	logger := log.TestingLogger()
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
			//stakesim.SimulateMsgCreateValidator(app.accountMapper, app.stakeKeeper),
			//stakesim.SimulateMsgEditValidator(app.stakeKeeper),
			//stakesim.SimulateMsgDelegate(app.accountMapper, app.stakeKeeper),
			//stakesim.SimulateMsgBeginUnbonding(app.accountMapper, app.stakeKeeper),
			//stakesim.SimulateMsgCompleteUnbonding(app.stakeKeeper),
			//stakesim.SimulateMsgBeginRedelegate(app.accountMapper, app.stakeKeeper),
			//stakesim.SimulateMsgCompleteRedelegate(app.stakeKeeper),
			slashingsim.SimulateMsgUnrevoke(app.slashingKeeper),
		},
		[]simulation.RandSetup{},
		[]simulation.Invariant{
			//simulation.PeriodicInvariant(allInvariants, 50, 0),
			allInvariants,
		},
		numKeys,
		numBlocks,
		blockSize,
		minTimePerBlock,
		maxTimePerBlock,
		signingFraction,
		evidenceFraction,
	)

}
