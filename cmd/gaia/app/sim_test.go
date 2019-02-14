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
	"github.com/tendermint/tendermint/crypto/secp256k1"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authsim "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banksim "github.com/cosmos/cosmos-sdk/x/bank/simulation"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrsim "github.com/cosmos/cosmos-sdk/x/distribution/simulation"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govsim "github.com/cosmos/cosmos-sdk/x/gov/simulation"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingsim "github.com/cosmos/cosmos-sdk/x/slashing/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingsim "github.com/cosmos/cosmos-sdk/x/staking/simulation"
)

var (
	genesisFile string
	seed        int64
	numBlocks   int
	blockSize   int
	enabled     bool
	verbose     bool
	commit      bool
	period      int
)

func init() {
	flag.StringVar(&genesisFile, "SimulationGenesis", "", "Custom simulation genesis file")
	flag.Int64Var(&seed, "SimulationSeed", 42, "Simulation random seed")
	flag.IntVar(&numBlocks, "SimulationNumBlocks", 500, "Number of blocks")
	flag.IntVar(&blockSize, "SimulationBlockSize", 200, "Operations per block")
	flag.BoolVar(&enabled, "SimulationEnabled", false, "Enable the simulation")
	flag.BoolVar(&verbose, "SimulationVerbose", false, "Verbose log output")
	flag.BoolVar(&commit, "SimulationCommit", false, "Have the simulation commit")
	flag.IntVar(&period, "SimulationPeriod", 1, "Run slow invariants only once every period assertions")
}

func appStateFromGenesisFileFn(r *rand.Rand, accs []simulation.Account, genesisTimestamp time.Time) (json.RawMessage, []simulation.Account, string) {
	var genesis tmtypes.GenesisDoc
	cdc := MakeCodec()
	bytes, err := ioutil.ReadFile(genesisFile)
	if err != nil {
		panic(err)
	}
	cdc.MustUnmarshalJSON(bytes, &genesis)
	var appState GenesisState
	cdc.MustUnmarshalJSON(genesis.AppState, &appState)
	var newAccs []simulation.Account
	for _, acc := range appState.Accounts {
		// Pick a random private key, since we don't know the actual key
		// This should be fine as it's only used for mock Tendermint validators
		// and these keys are never actually used to sign by mock Tendermint.
		privkeySeed := make([]byte, 15)
		r.Read(privkeySeed)
		privKey := secp256k1.GenPrivKeySecp256k1(privkeySeed)
		newAccs = append(newAccs, simulation.Account{privKey, privKey.PubKey(), acc.Address})
	}
	return genesis.AppState, newAccs, genesis.ChainID
}

func appStateRandomizedFn(r *rand.Rand, accs []simulation.Account, genesisTimestamp time.Time) (json.RawMessage, []simulation.Account, string) {

	var genesisAccounts []GenesisAccount

	amount := int64(r.Intn(1e12))
	numInitiallyBonded := int64(r.Intn(250))
	numAccs := int64(len(accs))
	if numInitiallyBonded > numAccs {
		numInitiallyBonded = numAccs
	}
	fmt.Printf("Selected randomly generated parameters for simulated genesis:\n"+
		"\t{amount of steak per account: %v, initially bonded validators: %v}\n",
		amount, numInitiallyBonded)

	// randomly generate some genesis accounts
	for i, acc := range accs {
		coins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(amount))}
		bacc := auth.NewBaseAccountWithAddress(acc.Address)
		bacc.SetCoins(coins)

		var gacc GenesisAccount

		// Only consider making a vesting account once the initial bonded validator
		// set is exhausted due to needing to track DelegatedVesting.
		if int64(i) > numInitiallyBonded && r.Intn(100) < 50 {
			var (
				vacc    auth.VestingAccount
				endTime int
			)

			startTime := genesisTimestamp.Unix()

			// Allow for some vesting accounts to vest very quickly while others very
			// slowly.
			if r.Intn(100) < 50 {
				endTime = randIntBetween(r, int(startTime), int(startTime+(60*60*24*30)))
			} else {
				endTime = randIntBetween(r, int(startTime), int(startTime+(60*60*12)))
			}

			if r.Intn(100) < 50 {
				vacc = auth.NewContinuousVestingAccount(&bacc, startTime, int64(endTime))
			} else {
				vacc = auth.NewDelayedVestingAccount(&bacc, int64(endTime))
			}

			gacc = NewGenesisAccountI(vacc)
		} else {
			gacc = NewGenesisAccount(&bacc)
		}

		genesisAccounts = append(genesisAccounts, gacc)
	}

	authGenesis := auth.GenesisState{
		Params: auth.Params{
			MaxMemoCharacters:      uint64(randIntBetween(r, 100, 200)),
			TxSigLimit:             uint64(r.Intn(7) + 1),
			TxSizeCostPerByte:      uint64(randIntBetween(r, 5, 15)),
			SigVerifyCostED25519:   uint64(randIntBetween(r, 500, 1000)),
			SigVerifyCostSecp256k1: uint64(randIntBetween(r, 500, 1000)),
		},
	}
	fmt.Printf("Selected randomly generated auth parameters:\n\t%+v\n", authGenesis)

	bankGenesis := bank.NewGenesisState(r.Int63n(2) == 0)
	fmt.Printf("Selected randomly generated bank parameters:\n\t%+v\n", bankGenesis)

	// Random genesis states
	vp := time.Duration(r.Intn(2*172800)) * time.Second
	govGenesis := gov.GenesisState{
		StartingProposalID: uint64(r.Intn(100)),
		DepositParams: gov.DepositParams{
			MinDeposit:       sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(r.Intn(1e3)))},
			MaxDepositPeriod: vp,
		},
		VotingParams: gov.VotingParams{
			VotingPeriod: vp,
		},
		TallyParams: gov.TallyParams{
			Threshold:         sdk.NewDecWithPrec(5, 1),
			Veto:              sdk.NewDecWithPrec(334, 3),
			GovernancePenalty: sdk.NewDecWithPrec(1, 2),
		},
	}
	fmt.Printf("Selected randomly generated governance parameters:\n\t%+v\n", govGenesis)

	stakingGenesis := staking.GenesisState{
		Pool: staking.InitialPool(),
		Params: staking.Params{
			UnbondingTime: time.Duration(randIntBetween(r, 60, 60*60*24*3*2)) * time.Second,
			MaxValidators: uint16(r.Intn(250)),
			BondDenom:     sdk.DefaultBondDenom,
		},
	}
	fmt.Printf("Selected randomly generated staking parameters:\n\t%+v\n", stakingGenesis)

	slashingGenesis := slashing.GenesisState{
		Params: slashing.Params{
			MaxEvidenceAge:          stakingGenesis.Params.UnbondingTime,
			SignedBlocksWindow:      int64(randIntBetween(r, 10, 1000)),
			MinSignedPerWindow:      sdk.NewDecWithPrec(int64(r.Intn(10)), 1),
			DowntimeJailDuration:    time.Duration(randIntBetween(r, 60, 60*60*24)) * time.Second,
			SlashFractionDoubleSign: sdk.NewDec(1).Quo(sdk.NewDec(int64(r.Intn(50) + 1))),
			SlashFractionDowntime:   sdk.NewDec(1).Quo(sdk.NewDec(int64(r.Intn(200) + 1))),
		},
	}
	fmt.Printf("Selected randomly generated slashing parameters:\n\t%+v\n", slashingGenesis)

	mintGenesis := mint.GenesisState{
		Minter: mint.InitialMinter(
			sdk.NewDecWithPrec(int64(r.Intn(99)), 2)),
		Params: mint.NewParams(
			sdk.DefaultBondDenom,
			sdk.NewDecWithPrec(int64(r.Intn(99)), 2),
			sdk.NewDecWithPrec(20, 2),
			sdk.NewDecWithPrec(7, 2),
			sdk.NewDecWithPrec(67, 2),
			uint64(60*60*8766/5)),
	}
	fmt.Printf("Selected randomly generated minting parameters:\n\t%+v\n", mintGenesis)

	var validators []staking.Validator
	var delegations []staking.Delegation

	valAddrs := make([]sdk.ValAddress, numInitiallyBonded)
	for i := 0; i < int(numInitiallyBonded); i++ {
		valAddr := sdk.ValAddress(accs[i].Address)
		valAddrs[i] = valAddr

		validator := staking.NewValidator(valAddr, accs[i].PubKey, staking.Description{})
		validator.Tokens = sdk.NewInt(amount)
		validator.DelegatorShares = sdk.NewDec(amount)
		delegation := staking.Delegation{accs[i].Address, valAddr, sdk.NewDec(amount)}
		validators = append(validators, validator)
		delegations = append(delegations, delegation)
	}

	stakingGenesis.Pool.NotBondedTokens = sdk.NewInt((amount * numAccs) + (numInitiallyBonded * amount))
	stakingGenesis.Validators = validators
	stakingGenesis.Delegations = delegations

	distrGenesis := distr.GenesisState{
		FeePool:             distr.InitialFeePool(),
		CommunityTax:        sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(30)), 2)),
		BaseProposerReward:  sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(30)), 2)),
		BonusProposerReward: sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(30)), 2)),
	}
	fmt.Printf("Selected randomly generated distribution parameters:\n\t%+v\n", distrGenesis)

	genesis := GenesisState{
		Accounts:     genesisAccounts,
		AuthData:     authGenesis,
		BankData:     bankGenesis,
		StakingData:  stakingGenesis,
		MintData:     mintGenesis,
		DistrData:    distrGenesis,
		SlashingData: slashingGenesis,
		GovData:      govGenesis,
	}

	// Marshal genesis
	appState, err := MakeCodec().MarshalJSON(genesis)
	if err != nil {
		panic(err)
	}

	return appState, accs, "simulation"
}

func appStateFn(r *rand.Rand, accs []simulation.Account, genesisTimestamp time.Time) (json.RawMessage, []simulation.Account, string) {
	if genesisFile != "" {
		return appStateFromGenesisFileFn(r, accs, genesisTimestamp)
	}
	return appStateRandomizedFn(r, accs, genesisTimestamp)
}

func randIntBetween(r *rand.Rand, min, max int) int {
	return r.Intn(max-min) + min
}

func testAndRunTxs(app *GaiaApp) []simulation.WeightedOperation {
	return []simulation.WeightedOperation{
		{5, authsim.SimulateDeductFee(app.accountKeeper, app.feeCollectionKeeper)},
		{100, banksim.SendMsg(app.accountKeeper, app.bankKeeper)},
		{10, banksim.SingleInputMsgMultiSend(app.accountKeeper, app.bankKeeper)},
		{50, distrsim.SimulateMsgSetWithdrawAddress(app.accountKeeper, app.distrKeeper)},
		{50, distrsim.SimulateMsgWithdrawDelegatorReward(app.accountKeeper, app.distrKeeper)},
		{50, distrsim.SimulateMsgWithdrawValidatorCommission(app.accountKeeper, app.distrKeeper)},
		{5, govsim.SimulateSubmittingVotingAndSlashingForProposal(app.govKeeper)},
		{100, govsim.SimulateMsgDeposit(app.govKeeper)},
		{100, stakingsim.SimulateMsgCreateValidator(app.accountKeeper, app.stakingKeeper)},
		{5, stakingsim.SimulateMsgEditValidator(app.stakingKeeper)},
		{100, stakingsim.SimulateMsgDelegate(app.accountKeeper, app.stakingKeeper)},
		{100, stakingsim.SimulateMsgUndelegate(app.accountKeeper, app.stakingKeeper)},
		{100, stakingsim.SimulateMsgBeginRedelegate(app.accountKeeper, app.stakingKeeper)},
		{100, slashingsim.SimulateMsgUnjail(app.slashingKeeper)},
	}
}

func invariants(app *GaiaApp) []sdk.Invariant {
	return []sdk.Invariant{
		simulation.PeriodicInvariant(banksim.NonnegativeBalanceInvariant(app.accountKeeper), period, 0),
		simulation.PeriodicInvariant(govsim.AllInvariants(), period, 0),
		simulation.PeriodicInvariant(distrsim.AllInvariants(app.distrKeeper, app.stakingKeeper), period, 0),
		simulation.PeriodicInvariant(stakingsim.AllInvariants(app.stakingKeeper, app.feeCollectionKeeper,
			app.distrKeeper, app.accountKeeper), period, 0),
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
	logger := log.NewNopLogger()

	var db dbm.DB
	dir, _ := ioutil.TempDir("", "goleveldb-gaia-sim")
	db, _ = dbm.NewGoLevelDB("Simulation", dir)
	defer func() {
		db.Close()
		os.RemoveAll(dir)
	}()
	app := NewGaiaApp(logger, db, nil, true)

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
	app := NewGaiaApp(logger, db, nil, true, fauxMerkleModeOpt)
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
	app := NewGaiaApp(logger, db, nil, true, fauxMerkleModeOpt)
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

	appState, _, err := app.ExportAppStateAndValidators(false, []string{})
	require.NoError(t, err)
	fmt.Printf("Importing genesis...\n")

	newDir, _ := ioutil.TempDir("", "goleveldb-gaia-sim-2")
	newDB, _ := dbm.NewGoLevelDB("Simulation-2", dir)
	defer func() {
		newDB.Close()
		os.RemoveAll(newDir)
	}()
	newApp := NewGaiaApp(log.NewNopLogger(), newDB, nil, true, fauxMerkleModeOpt)
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
		{app.keyStaking, newApp.keyStaking, [][]byte{staking.UnbondingQueueKey, staking.RedelegationQueueKey, staking.ValidatorQueueKey}}, // ordering may change but it doesn't matter
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
		require.True(t, equal,
			"unequal stores: %s / %s:\nstore A %X => %X\nstore B %X => %X",
			storeKeyA, storeKeyB, kvA.Key, kvA.Value, kvB.Key, kvB.Value,
		)
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
	app := NewGaiaApp(logger, db, nil, true, fauxMerkleModeOpt)
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

	appState, _, err := app.ExportAppStateAndValidators(true, []string{})
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
	newApp := NewGaiaApp(log.NewNopLogger(), newDB, nil, true, fauxMerkleModeOpt)
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
			app := NewGaiaApp(logger, db, nil, true)

			// Run randomized simulation
			simulation.SimulateFromSeed(
				t, app.BaseApp, appStateFn, seed,
				testAndRunTxs(app),
				[]sdk.Invariant{},
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
