package simapp

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
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
	"github.com/cosmos/cosmos-sdk/x/auth/genaccounts"
	authsim "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banksim "github.com/cosmos/cosmos-sdk/x/bank/simulation"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrsim "github.com/cosmos/cosmos-sdk/x/distribution/simulation"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govsim "github.com/cosmos/cosmos-sdk/x/gov/simulation"
	"github.com/cosmos/cosmos-sdk/x/mint"
	paramsim "github.com/cosmos/cosmos-sdk/x/params/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
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
	lean        bool
	commit      bool
	period      int
	onOperation bool // TODO Remove in favor of binary search for invariant violation
)

func init() {
	flag.StringVar(&genesisFile, "SimulationGenesis", "", "custom simulation genesis file")
	flag.Int64Var(&seed, "SimulationSeed", 42, "simulation random seed")
	flag.IntVar(&numBlocks, "SimulationNumBlocks", 500, "number of blocks")
	flag.IntVar(&blockSize, "SimulationBlockSize", 200, "operations per block")
	flag.BoolVar(&enabled, "SimulationEnabled", false, "enable the simulation")
	flag.BoolVar(&verbose, "SimulationVerbose", false, "verbose log output")
	flag.BoolVar(&lean, "SimulationLean", false, "lean simulation log output")
	flag.BoolVar(&commit, "SimulationCommit", false, "have the simulation commit")
	flag.IntVar(&period, "SimulationPeriod", 1, "run slow invariants only once every period assertions")
	flag.BoolVar(&onOperation, "SimulateEveryOperation", false, "run slow invariants every operation")
}

// helper function for populating input for SimulateFromSeed
func getSimulateFromSeedInput(tb testing.TB, w io.Writer, app *SimApp) (
	testing.TB, io.Writer, *baseapp.BaseApp, simulation.AppStateFn, int64,
	simulation.WeightedOperations, sdk.Invariants, int, int, bool, bool, bool) {

	return tb, w, app.BaseApp, appStateFn, seed,
		testAndRunTxs(app), invariants(app), numBlocks, blockSize, commit, lean, onOperation
}

func appStateFromGenesisFileFn(r *rand.Rand, accs []simulation.Account, genesisTimestamp time.Time,
) (json.RawMessage, []simulation.Account, string) {

	var genesis tmtypes.GenesisDoc
	cdc := MakeCodec()
	bytes, err := ioutil.ReadFile(genesisFile)
	if err != nil {
		panic(err)
	}
	cdc.MustUnmarshalJSON(bytes, &genesis)
	var appState GenesisState
	cdc.MustUnmarshalJSON(genesis.AppState, &appState)
	accounts := genaccounts.GetGenesisStateFromAppState(cdc, appState)

	var newAccs []simulation.Account
	for _, acc := range accounts {
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

// TODO refactor out random initialization code to the modules
func appStateRandomizedFn(r *rand.Rand, accs []simulation.Account, genesisTimestamp time.Time,
) (json.RawMessage, []simulation.Account, string) {

	var genesisAccounts []genaccounts.GenesisAccount
	genesisState := NewDefaultGenesisState()
	cdc := MakeCodec()

	amount := int64(r.Intn(1e12))
	numInitiallyBonded := int64(r.Intn(250))
	numAccs := int64(len(accs))
	if numInitiallyBonded > numAccs {
		numInitiallyBonded = numAccs
	}
	fmt.Printf("Selected randomly generated parameters for simulated genesis:\n"+
		"\t{amount of stake per account: %v, initially bonded validators: %v}\n",
		amount, numInitiallyBonded)

	// randomly generate some genesis accounts
	for i, acc := range accs {
		coins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(amount))}
		bacc := auth.NewBaseAccountWithAddress(acc.Address)
		bacc.SetCoins(coins)

		var gacc genaccounts.GenesisAccount

		// Only consider making a vesting account once the initial bonded validator
		// set is exhausted due to needing to track DelegatedVesting.
		if int64(i) > numInitiallyBonded && r.Intn(100) < 50 {
			var (
				vacc    auth.VestingAccount
				endTime int64
			)

			startTime := genesisTimestamp.Unix()

			// Allow for some vesting accounts to vest very quickly while others very
			// slowly.
			if r.Intn(100) < 50 {
				endTime = int64(simulation.RandIntBetween(r, int(startTime), int(startTime+(60*60*24*30))))
			} else {
				endTime = int64(simulation.RandIntBetween(r, int(startTime), int(startTime+(60*60*12))))
			}

			if startTime == endTime {
				endTime++
			}

			if r.Intn(100) < 50 {
				vacc = auth.NewContinuousVestingAccount(&bacc, startTime, endTime)
			} else {
				vacc = auth.NewDelayedVestingAccount(&bacc, endTime)
			}

			var err error
			gacc, err = genaccounts.NewGenesisAccountI(vacc)
			if err != nil {
				panic(err)
			}
		} else {
			gacc = genaccounts.NewGenesisAccount(&bacc)
		}

		genesisAccounts = append(genesisAccounts, gacc)
	}

	genesisState[genaccounts.ModuleName] = cdc.MustMarshalJSON(genesisAccounts)

	authGenesis := auth.NewGenesisState(
		nil,
		auth.NewParams(
			simulation.ModuleParamSimulator["MaxMemoCharacters"](r).(uint64),
			simulation.ModuleParamSimulator["TxSigLimit"](r).(uint64),
			simulation.ModuleParamSimulator["TxSizeCostPerByte"](r).(uint64),
			simulation.ModuleParamSimulator["SigVerifyCostED25519"](r).(uint64),
			simulation.ModuleParamSimulator["SigVerifyCostSecp256k1"](r).(uint64),
		),
	)
	fmt.Printf("Selected randomly generated auth parameters:\n\t%+v\n", authGenesis)
	genesisState[auth.ModuleName] = cdc.MustMarshalJSON(authGenesis)

	bankGenesis := bank.NewGenesisState(r.Int63n(2) == 0)
	genesisState[bank.ModuleName] = cdc.MustMarshalJSON(bankGenesis)
	fmt.Printf("Selected randomly generated bank parameters:\n\t%+v\n", bankGenesis)

	// Random genesis states
	vp := simulation.ModuleParamSimulator["VotingParams/VotingPeriod"](r).(time.Duration)
	govGenesis := gov.NewGenesisState(
		uint64(r.Intn(100)),
		gov.NewDepositParams(
			simulation.ModuleParamSimulator["DepositParams/MinDeposit"](r).(sdk.Coins),
			vp,
		),
		gov.NewVotingParams(vp),
		gov.NewTallyParams(
			simulation.ModuleParamSimulator["TallyParams/Quorum"](r).(sdk.Dec),
			simulation.ModuleParamSimulator["TallyParams/Threshold"](r).(sdk.Dec),
			simulation.ModuleParamSimulator["TallyParams/Veto"](r).(sdk.Dec),
		),
	)
	genesisState[gov.ModuleName] = cdc.MustMarshalJSON(govGenesis)
	fmt.Printf("Selected randomly generated governance parameters:\n\t%+v\n", govGenesis)

	stakingGenesis := staking.NewGenesisState(
		staking.InitialPool(),
		staking.NewParams(
			simulation.ModuleParamSimulator["UnbondingTime"](r).(time.Duration),
			simulation.ModuleParamSimulator["MaxValidators"](r).(uint16),
			7,
			sdk.DefaultBondDenom,
		),
		nil,
		nil,
	)
	fmt.Printf("Selected randomly generated staking parameters:\n\t%+v\n", stakingGenesis)

	slashingParams := slashing.NewParams(
		stakingGenesis.Params.UnbondingTime,
		simulation.ModuleParamSimulator["SignedBlocksWindow"](r).(int64),
		simulation.ModuleParamSimulator["MinSignedPerWindow"](r).(sdk.Dec),
		simulation.ModuleParamSimulator["DowntimeJailDuration"](r).(time.Duration),
		simulation.ModuleParamSimulator["SlashFractionDoubleSign"](r).(sdk.Dec),
		simulation.ModuleParamSimulator["SlashFractionDowntime"](r).(sdk.Dec),
	)
	slashingGenesis := slashing.NewGenesisState(slashingParams, nil, nil)
	genesisState[slashing.ModuleName] = cdc.MustMarshalJSON(slashingGenesis)
	fmt.Printf("Selected randomly generated slashing parameters:\n\t%+v\n", slashingGenesis)

	mintGenesis := mint.NewGenesisState(
		mint.InitialMinter(
			sdk.NewDecWithPrec(int64(r.Intn(99)), 2)),
		mint.NewParams(
			sdk.DefaultBondDenom,
			simulation.ModuleParamSimulator["InflationRateChange"](r).(sdk.Dec),
			simulation.ModuleParamSimulator["InflationMax"](r).(sdk.Dec),
			simulation.ModuleParamSimulator["InflationMin"](r).(sdk.Dec),
			simulation.ModuleParamSimulator["GoalBonded"](r).(sdk.Dec),
			uint64(60*60*8766/5),
		),
	)
	genesisState[mint.ModuleName] = cdc.MustMarshalJSON(mintGenesis)
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
	genesisState[staking.ModuleName] = cdc.MustMarshalJSON(stakingGenesis)

	// TODO make use NewGenesisState
	distrGenesis := distr.GenesisState{
		FeePool:             distr.InitialFeePool(),
		CommunityTax:        sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(30)), 2)),
		BaseProposerReward:  sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(30)), 2)),
		BonusProposerReward: sdk.NewDecWithPrec(1, 2).Add(sdk.NewDecWithPrec(int64(r.Intn(30)), 2)),
	}
	genesisState[distr.ModuleName] = cdc.MustMarshalJSON(distrGenesis)
	fmt.Printf("Selected randomly generated distribution parameters:\n\t%+v\n", distrGenesis)

	// Marshal genesis
	appState, err := MakeCodec().MarshalJSON(genesisState)
	if err != nil {
		panic(err)
	}

	return appState, accs, "simulation"
}

func appStateFn(r *rand.Rand, accs []simulation.Account, genesisTimestamp time.Time,
) (json.RawMessage, []simulation.Account, string) {

	if genesisFile != "" {
		return appStateFromGenesisFileFn(r, accs, genesisTimestamp)
	}
	return appStateRandomizedFn(r, accs, genesisTimestamp)
}

func testAndRunTxs(app *SimApp) []simulation.WeightedOperation {
	return []simulation.WeightedOperation{
		{5, authsim.SimulateDeductFee(app.accountKeeper, app.feeCollectionKeeper)},
		{100, banksim.SimulateMsgSend(app.accountKeeper, app.bankKeeper)},
		{10, banksim.SimulateSingleInputMsgMultiSend(app.accountKeeper, app.bankKeeper)},
		{50, distrsim.SimulateMsgSetWithdrawAddress(app.accountKeeper, app.distrKeeper)},
		{50, distrsim.SimulateMsgWithdrawDelegatorReward(app.accountKeeper, app.distrKeeper)},
		{50, distrsim.SimulateMsgWithdrawValidatorCommission(app.accountKeeper, app.distrKeeper)},
		{5, govsim.SimulateSubmittingVotingAndSlashingForProposal(app.govKeeper, govsim.SimulateTextProposalContent)},
		{5, govsim.SimulateSubmittingVotingAndSlashingForProposal(app.govKeeper, distrsim.SimulateCommunityPoolSpendProposalContent(app.distrKeeper))},
		{5, govsim.SimulateSubmittingVotingAndSlashingForProposal(app.govKeeper, paramsim.SimulateParamChangeProposalContent)},
		{100, govsim.SimulateMsgDeposit(app.govKeeper)},
		{100, stakingsim.SimulateMsgCreateValidator(app.accountKeeper, app.stakingKeeper)},
		{5, stakingsim.SimulateMsgEditValidator(app.stakingKeeper)},
		{100, stakingsim.SimulateMsgDelegate(app.accountKeeper, app.stakingKeeper)},
		{100, stakingsim.SimulateMsgUndelegate(app.accountKeeper, app.stakingKeeper)},
		{100, stakingsim.SimulateMsgBeginRedelegate(app.accountKeeper, app.stakingKeeper)},
		{100, slashingsim.SimulateMsgUnjail(app.slashingKeeper)},
	}
}

func invariants(app *SimApp) []sdk.Invariant {
	return simulation.PeriodicInvariants(app.crisisKeeper.Invariants(), period, 0)
}

// Pass this in as an option to use a dbStoreAdapter instead of an IAVLStore for simulation speed.
func fauxMerkleModeOpt(bapp *baseapp.BaseApp) {
	bapp.SetFauxMerkleMode()
}

// Profile with:
// /usr/local/go/bin/go test -benchmem -run=^$ github.com/cosmos/cosmos-sdk/simapp -bench ^BenchmarkFullAppSimulation$ -SimulationCommit=true -cpuprofile cpu.out
func BenchmarkFullAppSimulation(b *testing.B) {
	logger := log.NewNopLogger()

	var db dbm.DB
	dir, _ := ioutil.TempDir("", "goleveldb-app-sim")
	db, _ = sdk.NewLevelDB("Simulation", dir)
	defer func() {
		db.Close()
		os.RemoveAll(dir)
	}()
	app := NewSimApp(logger, db, nil, true, 0)

	// Run randomized simulation
	// TODO parameterize numbers, save for a later PR
	_, err := simulation.SimulateFromSeed(getSimulateFromSeedInput(b, os.Stdout, app))
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

func TestFullAppSimulation(t *testing.T) {
	if !enabled {
		t.Skip("Skipping application simulation")
	}

	var logger log.Logger

	if verbose {
		logger = log.TestingLogger()
	} else {
		logger = log.NewNopLogger()
	}

	var db dbm.DB
	dir, _ := ioutil.TempDir("", "goleveldb-app-sim")
	db, _ = sdk.NewLevelDB("Simulation", dir)

	defer func() {
		db.Close()
		os.RemoveAll(dir)
	}()

	app := NewSimApp(logger, db, nil, true, 0, fauxMerkleModeOpt)
	require.Equal(t, "SimApp", app.Name())

	// Run randomized simulation
	_, err := simulation.SimulateFromSeed(getSimulateFromSeedInput(t, os.Stdout, app))
	if commit {
		// for memdb:
		// fmt.Println("Database Size", db.Stats()["database.size"])
		fmt.Println("GoLevelDB Stats")
		fmt.Println(db.Stats()["leveldb.stats"])
		fmt.Println("GoLevelDB cached block size", db.Stats()["leveldb.cachedblock"])
	}

	require.Nil(t, err)
}

func TestAppImportExport(t *testing.T) {
	if !enabled {
		t.Skip("Skipping application import/export simulation")
	}

	var logger log.Logger
	if verbose {
		logger = log.TestingLogger()
	} else {
		logger = log.NewNopLogger()
	}

	var db dbm.DB
	dir, _ := ioutil.TempDir("", "goleveldb-app-sim")
	db, _ = sdk.NewLevelDB("Simulation", dir)

	defer func() {
		db.Close()
		os.RemoveAll(dir)
	}()

	app := NewSimApp(logger, db, nil, true, 0, fauxMerkleModeOpt)
	require.Equal(t, "SimApp", app.Name())

	// Run randomized simulation
	_, err := simulation.SimulateFromSeed(getSimulateFromSeedInput(t, os.Stdout, app))

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

	newDir, _ := ioutil.TempDir("", "goleveldb-app-sim-2")
	newDB, _ := sdk.NewLevelDB("Simulation-2", dir)

	defer func() {
		newDB.Close()
		os.RemoveAll(newDir)
	}()

	newApp := NewSimApp(log.NewNopLogger(), newDB, nil, true, 0, fauxMerkleModeOpt)
	require.Equal(t, "SimApp", newApp.Name())

	var genesisState GenesisState
	err = app.cdc.UnmarshalJSON(appState, &genesisState)
	if err != nil {
		panic(err)
	}

	ctxB := newApp.NewContext(true, abci.Header{})
	newApp.mm.InitGenesis(ctxB, genesisState)

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
		{app.keyStaking, newApp.keyStaking, [][]byte{staking.UnbondingQueueKey,
			staking.RedelegationQueueKey, staking.ValidatorQueueKey}}, // ordering may change but it doesn't matter
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

func TestAppSimulationAfterImport(t *testing.T) {
	if !enabled {
		t.Skip("Skipping application simulation after import")
	}

	var logger log.Logger
	if verbose {
		logger = log.TestingLogger()
	} else {
		logger = log.NewNopLogger()
	}

	dir, _ := ioutil.TempDir("", "goleveldb-app-sim")
	db, _ := sdk.NewLevelDB("Simulation", dir)

	defer func() {
		db.Close()
		os.RemoveAll(dir)
	}()

	app := NewSimApp(logger, db, nil, true, 0, fauxMerkleModeOpt)
	require.Equal(t, "SimApp", app.Name())

	// Run randomized simulation
	stopEarly, err := simulation.SimulateFromSeed(getSimulateFromSeedInput(t, os.Stdout, app))

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

	newDir, _ := ioutil.TempDir("", "goleveldb-app-sim-2")
	newDB, _ := sdk.NewLevelDB("Simulation-2", dir)

	defer func() {
		newDB.Close()
		os.RemoveAll(newDir)
	}()

	newApp := NewSimApp(log.NewNopLogger(), newDB, nil, true, 0, fauxMerkleModeOpt)
	require.Equal(t, "SimApp", newApp.Name())
	newApp.InitChain(abci.RequestInitChain{
		AppStateBytes: appState,
	})

	// Run randomized simulation on imported app
	_, err = simulation.SimulateFromSeed(getSimulateFromSeedInput(t, os.Stdout, newApp))
	require.Nil(t, err)
}

// TODO: Make another test for the fuzzer itself, which just has noOp txs
// and doesn't depend on the application.
func TestAppStateDeterminism(t *testing.T) {
	if !enabled {
		t.Skip("Skipping application simulation")
	}

	numSeeds := 3
	numTimesToRunPerSeed := 5
	appHashList := make([]json.RawMessage, numTimesToRunPerSeed)

	for i := 0; i < numSeeds; i++ {
		seed := rand.Int63()
		for j := 0; j < numTimesToRunPerSeed; j++ {
			logger := log.NewNopLogger()
			db := dbm.NewMemDB()
			app := NewSimApp(logger, db, nil, true, 0)

			// Run randomized simulation
			simulation.SimulateFromSeed(
				t, os.Stdout, app.BaseApp, appStateFn, seed,
				testAndRunTxs(app),
				[]sdk.Invariant{},
				50,
				100,
				true,
				false,
				false,
			)
			appHash := app.LastCommitID().Hash
			appHashList[j] = appHash
		}
		for k := 1; k < numTimesToRunPerSeed; k++ {
			require.Equal(t, appHashList[0], appHashList[k], "appHash list: %v", appHashList)
		}
	}
}

func BenchmarkInvariants(b *testing.B) {
	logger := log.NewNopLogger()
	dir, _ := ioutil.TempDir("", "goleveldb-app-invariant-bench")
	db, _ := sdk.NewLevelDB("simulation", dir)

	defer func() {
		db.Close()
		os.RemoveAll(dir)
	}()

	app := NewSimApp(logger, db, nil, true, 0)

	// 2. Run parameterized simulation (w/o invariants)
	_, err := simulation.SimulateFromSeed(
		b, ioutil.Discard, app.BaseApp, appStateFn, seed, testAndRunTxs(app),
		[]sdk.Invariant{}, numBlocks, blockSize, commit, lean, onOperation,
	)
	if err != nil {
		fmt.Println(err)
		b.FailNow()
	}

	ctx := app.NewContext(true, abci.Header{Height: app.LastBlockHeight() + 1})

	// 3. Benchmark each invariant separately
	//
	// NOTE: We use the crisis keeper as it has all the invariants registered with
	// their respective metadata which makes it useful for testing/benchmarking.
	for _, cr := range app.crisisKeeper.Routes() {
		b.Run(fmt.Sprintf("%s/%s", cr.ModuleName, cr.Route), func(b *testing.B) {
			if err := cr.Invar(ctx); err != nil {
				fmt.Println(err)
				b.FailNow()
			}
		})
	}
}
