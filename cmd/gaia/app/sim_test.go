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

	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
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
	stakeTypes "github.com/cosmos/cosmos-sdk/x/stake/types"
)

var (
	seed      int64
	numBlocks int
	blockSize int
	enabled   bool
	verbose   bool
	commit    bool
	period    int
)

func init() {
	flag.Int64Var(&seed, "SimulationSeed", 42, "Simulation random seed")
	flag.IntVar(&numBlocks, "SimulationNumBlocks", 500, "Number of blocks")
	flag.IntVar(&blockSize, "SimulationBlockSize", 200, "Operations per block")
	flag.BoolVar(&enabled, "SimulationEnabled", false, "Enable the simulation")
	flag.BoolVar(&verbose, "SimulationVerbose", false, "Verbose log output")
	flag.BoolVar(&commit, "SimulationCommit", false, "Have the simulation commit")
	flag.IntVar(&period, "SimulationPeriod", 1, "Run slow invariants only once every period assertions")
}

func appStateFn(r *rand.Rand, accs []simulation.Account) json.RawMessage {
	var genesisAccounts []GenesisAccount

	amount := int64(r.Intn(1e6))
	numInitiallyBonded := int64(r.Intn(250))
	numAccs := int64(len(accs))
	if numInitiallyBonded > numAccs {
		numInitiallyBonded = numAccs
	}
	fmt.Printf("Selected randomly generated parameters for simulated genesis:\n"+
		"\t{amount of steak per account: %v, initially bonded validators: %v}\n",
		amount, numInitiallyBonded)

	// Randomly generate some genesis accounts
	for _, acc := range accs {
		coins := sdk.Coins{sdk.NewCoin(stakeTypes.DefaultBondDenom, sdk.NewInt(amount))}
		genesisAccounts = append(genesisAccounts, GenesisAccount{
			Address: acc.Address,
			Coins:   coins,
		})
	}

	// Random genesis states
	vp := time.Duration(r.Intn(2*172800)) * time.Second
	govGenesis := gov.GenesisState{
		StartingProposalID: uint64(r.Intn(100)),
		DepositParams: gov.DepositParams{
			MinDeposit:       sdk.Coins{sdk.NewInt64Coin(stakeTypes.DefaultBondDenom, int64(r.Intn(1e3)))},
			MaxDepositPeriod: vp,
		},
		VotingParams: gov.VotingParams{
			VotingPeriod: vp,
		},
		TallyParams: gov.TallyParams{
			Threshold: sdk.NewDecWithPrec(5, 1),
			Veto:      sdk.NewDecWithPrec(334, 3),
		},
	}
	fmt.Printf("Selected randomly generated governance parameters:\n\t%+v\n", govGenesis)

	stakeGenesis := stake.GenesisState{
		Pool: stake.InitialPool(),
		Params: stake.Params{
			UnbondingTime: time.Duration(r.Intn(60*60*24*3*2)) * time.Second,
			MaxValidators: uint16(r.Intn(250)),
			BondDenom:     stakeTypes.DefaultBondDenom,
		},
	}
	fmt.Printf("Selected randomly generated staking parameters:\n\t%+v\n", stakeGenesis)

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
	fmt.Printf("Selected randomly generated slashing parameters:\n\t%+v\n", slashingGenesis)

	mintGenesis := mint.GenesisState{
		Minter: mint.InitialMinter(
			sdk.NewDecWithPrec(int64(r.Intn(99)), 2)),
		Params: mint.NewParams(
			stakeTypes.DefaultBondDenom,
			sdk.NewDecWithPrec(int64(r.Intn(99)), 2),
			sdk.NewDecWithPrec(20, 2),
			sdk.NewDecWithPrec(7, 2),
			sdk.NewDecWithPrec(67, 2),
			uint64(60*60*8766/5)),
	}
	fmt.Printf("Selected randomly generated minting parameters:\n\t%+v\n", mintGenesis)

	var validators []stake.Validator
	var delegations []stake.Delegation

	valAddrs := make([]sdk.ValAddress, numInitiallyBonded)
	for i := 0; i < int(numInitiallyBonded); i++ {
		valAddr := sdk.ValAddress(accs[i].Address)
		valAddrs[i] = valAddr

		validator := stake.NewValidator(valAddr, accs[i].PubKey, stake.Description{})
		validator.Tokens = sdk.NewDec(amount)
		validator.DelegatorShares = sdk.NewDec(amount)
		delegation := stake.Delegation{accs[i].Address, valAddr, sdk.NewDec(amount)}
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
		simulation.PeriodicInvariant(banksim.NonnegativeBalanceInvariant(app.accountKeeper), period, 0),
		simulation.PeriodicInvariant(govsim.AllInvariants(), period, 0),
		simulation.PeriodicInvariant(distrsim.AllInvariants(app.distrKeeper, app.stakeKeeper), period, 0),
		simulation.PeriodicInvariant(stakesim.AllInvariants(app.bankKeeper, app.stakeKeeper,
			app.feeCollectionKeeper, app.distrKeeper, app.accountKeeper), period, 0),
		simulation.PeriodicInvariant(slashingsim.AllInvariants(), period, 0),
	}
}

// Pass this in as an option to use a dbStoreAdapter instead of an IAVLStore for simulation speed.
func fauxMerkleModeOpt(bapp *baseapp.BaseApp) {
	bapp.SetFauxMerkleMode()
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
	_, err := simulation.SimulateFromSeed(
		b, app.BaseApp, appStateFn, seed,
		testAndRunTxs(app),
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
	app := NewGaiaApp(logger, db, nil, fauxMerkleModeOpt)
	require.Equal(t, "GaiaApp", app.Name())

	// Run randomized simulation
	_, err := simulation.SimulateFromSeed(
		t, app.BaseApp, appStateFn, seed,
		testAndRunTxs(app),
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

func TestGaiaImportExport(t *testing.T) {
	if !enabled {
		t.Skip("Skipping Gaia import/export simulation")
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
	app := NewGaiaApp(logger, db, nil, fauxMerkleModeOpt)
	require.Equal(t, "GaiaApp", app.Name())

	// Run randomized simulation
	_, err := simulation.SimulateFromSeed(
		t, app.BaseApp, appStateFn, seed,
		testAndRunTxs(app),
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

	fmt.Printf("Exporting genesis...\n")

	appState, _, err := app.ExportAppStateAndValidators(false)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Importing genesis...\n")

	newDir, _ := ioutil.TempDir("", "goleveldb-gaia-sim-2")
	newDB, _ := dbm.NewGoLevelDB("Simulation-2", dir)
	defer func() {
		newDB.Close()
		os.RemoveAll(newDir)
	}()
	newApp := NewGaiaApp(log.NewNopLogger(), newDB, nil, fauxMerkleModeOpt)
	require.Equal(t, "GaiaApp", newApp.Name())
	var genesisState GenesisState
	err = app.cdc.UnmarshalJSON(appState, &genesisState)
	if err != nil {
		panic(err)
	}
	ctxB := newApp.NewContext(true, abci.Header{})
	newApp.initFromGenesisState(ctxB, genesisState)

	fmt.Printf("Comparing stores...\n")
	ctxA := app.NewContext(true, abci.Header{})
	type StoreKeysPrefixes struct {
		A        sdk.StoreKey
		B        sdk.StoreKey
		Prefixes [][]byte
	}
	storeKeysPrefixes := []StoreKeysPrefixes{
		{app.keyMain, newApp.keyMain, [][]byte{}},
		{app.keyAccount, newApp.keyAccount, [][]byte{}},
		{app.keyStake, newApp.keyStake, [][]byte{stake.UnbondingQueueKey, stake.RedelegationQueueKey, stake.ValidatorQueueKey}}, // ordering may change but it doesn't matter
		{app.keySlashing, newApp.keySlashing, [][]byte{}},
		{app.keyMint, newApp.keyMint, [][]byte{}},
		{app.keyDistr, newApp.keyDistr, [][]byte{}},
		{app.keyFeeCollection, newApp.keyFeeCollection, [][]byte{}},
		{app.keyParams, newApp.keyParams, [][]byte{}},
		{app.keyGov, newApp.keyGov, [][]byte{}},
	}
	for _, storeKeysPrefix := range storeKeysPrefixes {
		storeKeyA := storeKeysPrefix.A
		storeKeyB := storeKeysPrefix.B
		prefixes := storeKeysPrefix.Prefixes
		storeA := ctxA.KVStore(storeKeyA)
		storeB := ctxB.KVStore(storeKeyB)
		kvA, kvB, count, equal := sdk.DiffKVStores(storeA, storeB, prefixes)
		fmt.Printf("Compared %d key/value pairs between %s and %s\n", count, storeKeyA, storeKeyB)
		require.True(t, equal, "unequal stores: %s / %s:\nstore A %s (%X) => %s (%X)\nstore B %s (%X) => %s (%X)",
			storeKeyA, storeKeyB, kvA.Key, kvA.Key, kvA.Value, kvA.Value, kvB.Key, kvB.Key, kvB.Value, kvB.Value)
	}

}

func TestGaiaSimulationAfterImport(t *testing.T) {
	if !enabled {
		t.Skip("Skipping Gaia simulation after import")
	}

	// Setup Gaia application
	var logger log.Logger
	if verbose {
		logger = log.TestingLogger()
	} else {
		logger = log.NewNopLogger()
	}
	dir, _ := ioutil.TempDir("", "goleveldb-gaia-sim")
	db, _ := dbm.NewGoLevelDB("Simulation", dir)
	defer func() {
		db.Close()
		os.RemoveAll(dir)
	}()
	app := NewGaiaApp(logger, db, nil, fauxMerkleModeOpt)
	require.Equal(t, "GaiaApp", app.Name())

	// Run randomized simulation
	stopEarly, err := simulation.SimulateFromSeed(
		t, app.BaseApp, appStateFn, seed,
		testAndRunTxs(app),
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

	if stopEarly {
		// we can't export or import a zero-validator genesis
		fmt.Printf("We can't export or import a zero-validator genesis, exiting test...\n")
		return
	}

	fmt.Printf("Exporting genesis...\n")

	appState, _, err := app.ExportAppStateAndValidators(true)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Importing genesis...\n")

	newDir, _ := ioutil.TempDir("", "goleveldb-gaia-sim-2")
	newDB, _ := dbm.NewGoLevelDB("Simulation-2", dir)
	defer func() {
		newDB.Close()
		os.RemoveAll(newDir)
	}()
	newApp := NewGaiaApp(log.NewNopLogger(), newDB, nil, fauxMerkleModeOpt)
	require.Equal(t, "GaiaApp", newApp.Name())
	newApp.InitChain(abci.RequestInitChain{
		AppStateBytes: appState,
	})

	// Run randomized simulation on imported app
	_, err = simulation.SimulateFromSeed(
		t, newApp.BaseApp, appStateFn, seed,
		testAndRunTxs(newApp),
		invariants(newApp),
		numBlocks,
		blockSize,
		commit,
	)
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
				[]simulation.Invariant{},
				50,
				100,
				true,
			)
			appHash := app.LastCommitID().Hash
			appHashList[j] = appHash
		}
		for k := 1; k < numTimesToRunPerSeed; k++ {
			require.Equal(t, appHashList[0], appHashList[k], "appHash list: %v", appHashList)
		}
	}
}
