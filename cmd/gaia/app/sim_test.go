package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authsim "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	banksim "github.com/cosmos/cosmos-sdk/x/bank/simulation"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrsim "github.com/cosmos/cosmos-sdk/x/distribution/simulation"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govsim "github.com/cosmos/cosmos-sdk/x/gov/simulation"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing"
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
	commit    bool
)

func init() {
	flag.Int64Var(&seed, "SimulationSeed", 42, "Simulation random seed")
	flag.IntVar(&numBlocks, "SimulationNumBlocks", 500, "Number of blocks")
	flag.IntVar(&blockSize, "SimulationBlockSize", 200, "Operations per block")
	flag.BoolVar(&enabled, "SimulationEnabled", false, "Enable the simulation")
	flag.BoolVar(&verbose, "SimulationVerbose", false, "Verbose log output")
	flag.BoolVar(&commit, "SimulationCommit", false, "Have the simulation commit")
}

func appStateFn(r *rand.Rand, accs []simulation.Account) json.RawMessage {
	var genesisAccounts []GenesisAccount

	amount := int64(r.Intn(1e6))
	numInitiallyBonded := int64(r.Intn(250))
	numAccs := int64(len(accs))
	if numInitiallyBonded > numAccs {
		numInitiallyBonded = numAccs
	}
	fmt.Printf("Selected randomly generated parameters for simulated genesis: {amount of steak per account: %v, initially bonded validators: %v}\n", amount, numInitiallyBonded)

	// Randomly generate some genesis accounts
	for _, acc := range accs {
		coins := sdk.Coins{sdk.Coin{"steak", sdk.NewInt(amount)}}
		genesisAccounts = append(genesisAccounts, GenesisAccount{
			Address: acc.Address,
			Coins:   coins,
		})
	}

	// Random genesis states
	govGenesis := gov.GenesisState{
		StartingProposalID: uint64(r.Intn(100)),
		DepositParams: gov.DepositParams{
			MinDeposit:       sdk.Coins{sdk.NewInt64Coin("steak", int64(r.Intn(1e3)))},
			MaxDepositPeriod: time.Duration(r.Intn(2*172800)) * time.Second,
		},
		VotingParams: gov.VotingParams{
			VotingPeriod: time.Duration(r.Intn(2*172800)) * time.Second,
		},
		TallyParams: gov.TallyParams{
			Threshold:         sdk.NewDecWithPrec(5, 1),
			Veto:              sdk.NewDecWithPrec(334, 3),
			GovernancePenalty: sdk.NewDecWithPrec(1, 2),
		},
	}
	fmt.Printf("Selected randomly generated governance parameters: %+v\n", govGenesis)
	stakeGenesis := stake.GenesisState{
		Pool: stake.InitialPool(),
		Params: stake.Params{
			UnbondingTime: time.Duration(r.Intn(60*60*24*3*2)) * time.Second,
			MaxValidators: uint16(r.Intn(250)),
			BondDenom:     "steak",
		},
	}
	fmt.Printf("Selected randomly generated staking parameters: %+v\n", stakeGenesis)
	slashingGenesis := slashing.GenesisState{
		Params: slashing.Params{
			MaxEvidenceAge:           stakeGenesis.Params.UnbondingTime,
			DoubleSignUnbondDuration: time.Duration(r.Intn(60*60*24)) * time.Second,
			SignedBlocksWindow:       int64(r.Intn(1000)),
			DowntimeUnbondDuration:   time.Duration(r.Intn(86400)) * time.Second,
			MinSignedPerWindow:       sdk.NewDecWithPrec(int64(r.Intn(10)), 1),
			SlashFractionDoubleSign:  sdk.NewDec(1).Quo(sdk.NewDec(int64(r.Intn(50) + 1))),
			SlashFractionDowntime:    sdk.NewDec(1).Quo(sdk.NewDec(int64(r.Intn(200) + 1))),
		},
	}
	fmt.Printf("Selected randomly generated slashing parameters: %+v\n", slashingGenesis)
	mintGenesis := mint.GenesisState{
		Minter: mint.Minter{
			InflationLastTime: time.Unix(0, 0),
			Inflation:         sdk.NewDecWithPrec(int64(r.Intn(99)), 2),
		},
		Params: mint.Params{
			MintDenom:           "steak",
			InflationRateChange: sdk.NewDecWithPrec(int64(r.Intn(99)), 2),
			InflationMax:        sdk.NewDecWithPrec(20, 2),
			InflationMin:        sdk.NewDecWithPrec(7, 2),
			GoalBonded:          sdk.NewDecWithPrec(67, 2),
		},
	}
	fmt.Printf("Selected randomly generated minting parameters: %v\n", mintGenesis)
	var validators []stake.Validator
	var delegations []stake.Delegation

	valAddrs := make([]sdk.ValAddress, numInitiallyBonded)
	for i := 0; i < int(numInitiallyBonded); i++ {
		valAddr := sdk.ValAddress(accs[i].Address)
		valAddrs[i] = valAddr

		validator := stake.NewValidator(valAddr, accs[i].PubKey, stake.Description{})
		validator.Tokens = sdk.NewDec(amount)
		validator.DelegatorShares = sdk.NewDec(amount)
		delegation := stake.Delegation{accs[i].Address, valAddr, sdk.NewDec(amount), 0}
		validators = append(validators, validator)
		delegations = append(delegations, delegation)
	}
	stakeGenesis.Pool.LooseTokens = sdk.NewDec((amount * numAccs) + (numInitiallyBonded * amount))
	stakeGenesis.Validators = validators
	stakeGenesis.Bonds = delegations

	genesis := GenesisState{
		Accounts:     genesisAccounts,
		StakeData:    stakeGenesis,
		MintData:     mintGenesis,
		DistrData:    distr.DefaultGenesisWithValidators(valAddrs),
		SlashingData: slashingGenesis,
		GovData:      govGenesis,
	}

	// Marshal genesis
	appState, err := MakeCodec().MarshalJSON(genesis)
	if err != nil {
		panic(err)
	}

	return appState
}

func testAndRunTxs(app *GaiaApp) []simulation.WeightedOperation {
	return []simulation.WeightedOperation{
		{5, authsim.SimulateDeductFee(app.accountKeeper, app.feeCollectionKeeper)},
		{100, banksim.SingleInputSendMsg(app.accountKeeper, app.bankKeeper)},
		{50, distrsim.SimulateMsgSetWithdrawAddress(app.accountKeeper, app.distrKeeper)},
		{50, distrsim.SimulateMsgWithdrawDelegatorRewardsAll(app.accountKeeper, app.distrKeeper)},
		{50, distrsim.SimulateMsgWithdrawDelegatorReward(app.accountKeeper, app.distrKeeper)},
		{50, distrsim.SimulateMsgWithdrawValidatorRewardsAll(app.accountKeeper, app.distrKeeper)},
		{5, govsim.SimulateSubmittingVotingAndSlashingForProposal(app.govKeeper, app.stakeKeeper)},
		{100, govsim.SimulateMsgDeposit(app.govKeeper)},
		{100, stakesim.SimulateMsgCreateValidator(app.accountKeeper, app.stakeKeeper)},
		{5, stakesim.SimulateMsgEditValidator(app.stakeKeeper)},
		{100, stakesim.SimulateMsgDelegate(app.accountKeeper, app.stakeKeeper)},
		{100, stakesim.SimulateMsgBeginUnbonding(app.accountKeeper, app.stakeKeeper)},
		{100, stakesim.SimulateMsgBeginRedelegate(app.accountKeeper, app.stakeKeeper)},
		{100, slashingsim.SimulateMsgUnjail(app.slashingKeeper)},
	}
}

func invariants(app *GaiaApp) []simulation.Invariant {
	return []simulation.Invariant{
		banksim.NonnegativeBalanceInvariant(app.accountKeeper),
		govsim.AllInvariants(),
		distrsim.AllInvariants(app.distrKeeper, app.stakeKeeper),
		stakesim.AllInvariants(app.bankKeeper, app.stakeKeeper,
			app.feeCollectionKeeper, app.distrKeeper, app.accountKeeper),
		slashingsim.AllInvariants(),
	}
}

// Profile with:
// /usr/local/go/bin/go test -benchmem -run=^$ github.com/cosmos/cosmos-sdk/cmd/gaia/app -bench ^BenchmarkFullGaiaSimulation$ -SimulationCommit=true -cpuprofile cpu.out
func BenchmarkFullGaiaSimulation(b *testing.B) {
	// Setup Gaia application
	var logger log.Logger
	logger = log.NewNopLogger()
	var db dbm.DB
	dir, _ := ioutil.TempDir("", "goleveldb-gaia-sim")
	db, _ = dbm.NewGoLevelDB("Simulation", dir)
	defer func() {
		db.Close()
		os.RemoveAll(dir)
	}()
	app := NewGaiaApp(logger, db, nil)

	// Run randomized simulation
	// TODO parameterize numbers, save for a later PR
	err := simulation.SimulateFromSeed(
		b, app.BaseApp, appStateFn, seed,
		testAndRunTxs(app),
		[]simulation.RandSetup{},
		invariants(app), // these shouldn't get ran
		numBlocks,
		blockSize,
		commit,
	)
	if err != nil {
		fmt.Println(err)
		b.Fail()
	}
	if commit {
		fmt.Println("GoLevelDB Stats")
		fmt.Println(db.Stats()["leveldb.stats"])
		fmt.Println("GoLevelDB cached block size", db.Stats()["leveldb.cachedblock"])
	}
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
	var db dbm.DB
	dir, _ := ioutil.TempDir("", "goleveldb-gaia-sim")
	db, _ = dbm.NewGoLevelDB("Simulation", dir)
	defer func() {
		db.Close()
		os.RemoveAll(dir)
	}()
	app := NewGaiaApp(logger, db, nil)
	require.Equal(t, "GaiaApp", app.Name())

	// Run randomized simulation
	err := simulation.SimulateFromSeed(
		t, app.BaseApp, appStateFn, seed,
		testAndRunTxs(app),
		[]simulation.RandSetup{},
		invariants(app),
		numBlocks,
		blockSize,
		commit,
	)
	if commit {
		// for memdb:
		// fmt.Println("Database Size", db.Stats()["database.size"])
		fmt.Println("GoLevelDB Stats")
		fmt.Println(db.Stats()["leveldb.stats"])
		fmt.Println("GoLevelDB cached block size", db.Stats()["leveldb.cachedblock"])
	}
	require.Nil(t, err)
}

// TODO: Make another test for the fuzzer itself, which just has noOp txs
// and doesn't depend on gaia
func TestAppStateDeterminism(t *testing.T) {
	if !enabled {
		t.Skip("Skipping Gaia simulation")
	}

	numSeeds := 3
	numTimesToRunPerSeed := 5
	appHashList := make([]json.RawMessage, numTimesToRunPerSeed)

	for i := 0; i < numSeeds; i++ {
		seed := rand.Int63()
		for j := 0; j < numTimesToRunPerSeed; j++ {
			logger := log.NewNopLogger()
			db := dbm.NewMemDB()
			app := NewGaiaApp(logger, db, nil)

			// Run randomized simulation
			simulation.SimulateFromSeed(
				t, app.BaseApp, appStateFn, seed,
				testAndRunTxs(app),
				[]simulation.RandSetup{},
				[]simulation.Invariant{},
				50,
				100,
				true,
			)
			//app.Commit()
			appHash := app.LastCommitID().Hash
			appHashList[j] = appHash
		}
		for k := 1; k < numTimesToRunPerSeed; k++ {
			require.Equal(t, appHashList[0], appHashList[k], "appHash list: %v", appHashList)
		}
	}
}
