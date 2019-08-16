package simapp

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmtypes "github.com/tendermint/tendermint/types"

	authsim "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	banksim "github.com/cosmos/cosmos-sdk/x/bank/simulation"
	distrsim "github.com/cosmos/cosmos-sdk/x/distribution/simulation"
	"github.com/cosmos/cosmos-sdk/x/genaccounts"
	genaccsim "github.com/cosmos/cosmos-sdk/x/genaccounts/simulation"
	govsim "github.com/cosmos/cosmos-sdk/x/gov/simulation"
	mintsim "github.com/cosmos/cosmos-sdk/x/mint/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	slashingsim "github.com/cosmos/cosmos-sdk/x/slashing/simulation"
	stakingsim "github.com/cosmos/cosmos-sdk/x/staking/simulation"
	supplysim "github.com/cosmos/cosmos-sdk/x/supply/simulation"
)

// AppStateFn returns the initial application state using a genesis or the simulation parameters.
// It panics if the user provides files for both of them.
// If a file is not given for the genesis or the sim params, it creates a randomized one.
func AppStateFn(
	r *rand.Rand, accs []simulation.Account, config simulation.Config,
) (appState json.RawMessage, simAccs []simulation.Account, chainID string, genesisTimestamp time.Time) {

	cdc := MakeCodec()

	if flagGenesisTimeValue == 0 {
		genesisTimestamp = simulation.RandTimestamp(r)
	} else {
		genesisTimestamp = time.Unix(flagGenesisTimeValue, 0)
	}

	switch {
	case config.ParamsFile != "" && config.GenesisFile != "":
		panic("cannot provide both a genesis file and a params file")

	case config.GenesisFile != "":
		appState, simAccs, chainID = AppStateFromGenesisFileFn(r, config)

	case config.ParamsFile != "":
		appParams := make(simulation.AppParams)
		bz, err := ioutil.ReadFile(config.ParamsFile)
		if err != nil {
			panic(err)
		}

		cdc.MustUnmarshalJSON(bz, &appParams)
		appState, simAccs, chainID = AppStateRandomizedFn(r, accs, genesisTimestamp, appParams)

	default:
		appParams := make(simulation.AppParams)
		appState, simAccs, chainID = AppStateRandomizedFn(r, accs, genesisTimestamp, appParams)
	}

	return appState, simAccs, chainID, genesisTimestamp
}

// AppStateRandomizedFn creates calls each module's GenesisState generator function
// and creates
func AppStateRandomizedFn(
	r *rand.Rand, accs []simulation.Account, genesisTimestamp time.Time, appParams simulation.AppParams,
) (json.RawMessage, []simulation.Account, string) {

	cdc := MakeCodec()
	genesisState := NewDefaultGenesisState()

	var (
		amount             int64
		numInitiallyBonded int64
	)

	appParams.GetOrGenerate(cdc, StakePerAccount, &amount, r,
		func(r *rand.Rand) { amount = int64(r.Intn(1e12)) })
	appParams.GetOrGenerate(cdc, InitiallyBondedValidators, &amount, r,
		func(r *rand.Rand) { numInitiallyBonded = int64(r.Intn(250)) })

	numAccs := int64(len(accs))
	if numInitiallyBonded > numAccs {
		numInitiallyBonded = numAccs
	}

	fmt.Printf(
		`Selected randomly generated parameters for simulated genesis:
{
  stake_per_account: "%d",
  initially_bonded_validators: "%d"
}
`, amount, numInitiallyBonded,
	)

	genaccsim.RandomizedGenState(cdc, r, genesisState, accs, amount, numInitiallyBonded, genesisTimestamp)
	authsim.RandomizedGenState(cdc, r, genesisState)
	banksim.RandomizedGenState(cdc, r, genesisState)
	supplysim.RandomizedGenState(cdc, r, genesisState, accs, amount, numInitiallyBonded)
	govsim.RandomizedGenState(cdc, r, genesisState)
	mintsim.RandomizedGenState(cdc, r, genesisState)
	distrsim.RandomizedGenState(cdc, r, genesisState)
	stakingGen := stakingsim.RandomizedGenState(cdc, r, genesisState, accs, amount, numInitiallyBonded)
	slashingsim.RandomizedGenState(cdc, r, genesisState, stakingGen.Params.UnbondingTime)

	appState, err := cdc.MarshalJSON(genesisState)
	if err != nil {
		panic(err)
	}

	return appState, accs, "simulation"
}

// AppStateFromGenesisFileFn util function to generate the genesis AppState
// from a genesis.json file
func AppStateFromGenesisFileFn(r *rand.Rand, config simulation.Config) (json.RawMessage, []simulation.Account, string) {

	var genesis tmtypes.GenesisDoc
	cdc := MakeCodec()

	bytes, err := ioutil.ReadFile(config.GenesisFile)
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
