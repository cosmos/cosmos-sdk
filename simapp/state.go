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

	"github.com/cosmos/cosmos-sdk/x/genaccounts"
	nftsim "github.com/cosmos/cosmos-sdk/x/nft/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
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
  stake_per_account: "%v",
  initially_bonded_validators: "%v"
}
`, amount, numInitiallyBonded,
	)

	GenGenesisAccounts(cdc, r, accs, genesisTimestamp, amount, numInitiallyBonded, genesisState)
	GenAuthGenesisState(cdc, r, appParams, genesisState)
	GenBankGenesisState(cdc, r, appParams, genesisState)
	GenSupplyGenesisState(cdc, amount, numInitiallyBonded, int64(len(accs)), genesisState)
	GenGovGenesisState(cdc, r, appParams, genesisState)
	GenMintGenesisState(cdc, r, appParams, genesisState)
	nftsim.GenNFTGenesisState(cdc, r, accs, appParams, genesisState)
	GenDistrGenesisState(cdc, r, appParams, genesisState)
	stakingGen := GenStakingGenesisState(cdc, r, accs, amount, numAccs, numInitiallyBonded, appParams, genesisState)
	GenSlashingGenesisState(cdc, r, stakingGen, appParams, genesisState)

	appState, err := MakeCodec().MarshalJSON(genesisState)
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
		newAccs = append(newAccs, simulation.Account{PrivKey: privKey, PubKey: privKey.PubKey(), Address: acc.Address})
	}

	return genesis.AppState, newAccs, genesis.ChainID
}
