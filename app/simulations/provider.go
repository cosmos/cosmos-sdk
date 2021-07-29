package simulations

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"time"

	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/container"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type GenesisTime time.Time

type inputArgs struct {
	container.StructArgs

	AppModules  map[string]module.AppModule
	Codec       codec.Codec
	GenesisTime GenesisTime `optional:"true"`
}

var Provider = container.Options(
	container.Provide(
		func(inputs inputArgs) (
			func(*baseapp.BaseApp),
			*module.SimulationManager,
			simulation.AppStateFn,
		) {
			var simMods []module.AppModuleSimulation
			for _, mod := range inputs.AppModules {
				simMod, ok := mod.(module.AppModuleSimulation)
				if ok {
					simMods = append(simMods, simMod)
				}
			}

			sm := &module.SimulationManager{
				Modules:       simMods,
				StoreDecoders: nil,
			}
			sm.RegisterStoreDecoders()

			return func(app *baseapp.BaseApp) {

			}, sm, appStateFn(inputs.Codec, sm, inputs.GenesisTime, inputs.AppModules)
		},
	),
)

func appStateFn(
	cdc codec.JSONCodec,
	simManager *module.SimulationManager,
	genesisTimestampValue GenesisTime,
	modules map[string]module.AppModule,
) simtypes.AppStateFn {
	return func(r *rand.Rand, accs []simtypes.Account, config simtypes.Config,
	) (appState json.RawMessage, simAccs []simtypes.Account, chainID string, genesisTimestamp time.Time) {

		genesisTimestamp = time.Time(genesisTimestampValue)
		if genesisTimestamp.IsZero() {
			genesisTimestamp = simtypes.RandTimestamp(r)
		}

		chainID = config.ChainID
		switch {
		case config.ParamsFile != "" && config.GenesisFile != "":
			panic("cannot provide both a genesis file and a params file")

		case config.GenesisFile != "":
			// override the default chain-id from simapp to set it later to the config
			genesisDoc, accounts := AppStateFromGenesisFileFn(r, cdc, config.GenesisFile)

			if genesisTimestamp.IsZero() {
				// use genesis timestamp if no custom timestamp is provided (i.e no random timestamp)
				genesisTimestamp = genesisDoc.GenesisTime
			}

			appState = genesisDoc.AppState
			chainID = genesisDoc.ChainID
			simAccs = accounts

		case config.ParamsFile != "":
			appParams := make(simtypes.AppParams)
			bz, err := ioutil.ReadFile(config.ParamsFile)
			if err != nil {
				panic(err)
			}

			err = json.Unmarshal(bz, &appParams)
			if err != nil {
				panic(err)
			}
			appState, simAccs = AppStateRandomizedFn(simManager, r, cdc, accs, genesisTimestamp, appParams, modules)

		default:
			appParams := make(simtypes.AppParams)
			appState, simAccs = AppStateRandomizedFn(simManager, r, cdc, accs, genesisTimestamp, appParams, modules)
		}

		rawState := make(map[string]json.RawMessage)
		err := json.Unmarshal(appState, &rawState)
		if err != nil {
			panic(err)
		}

		stakingStateBz, ok := rawState[stakingtypes.ModuleName]
		if !ok {
			panic("staking genesis state is missing")
		}

		stakingState := new(stakingtypes.GenesisState)
		err = cdc.UnmarshalJSON(stakingStateBz, stakingState)
		if err != nil {
			panic(err)
		}
		// compute not bonded balance
		notBondedTokens := sdk.ZeroInt()
		for _, val := range stakingState.Validators {
			if val.Status != stakingtypes.Unbonded {
				continue
			}
			notBondedTokens = notBondedTokens.Add(val.GetTokens())
		}
		notBondedCoins := sdk.NewCoin(stakingState.Params.BondDenom, notBondedTokens)
		// edit bank state to make it have the not bonded pool tokens
		bankStateBz, ok := rawState[banktypes.ModuleName]
		// TODO(fdymylja/jonathan): should we panic in this case
		if !ok {
			panic("bank genesis state is missing")
		}
		bankState := new(banktypes.GenesisState)
		err = cdc.UnmarshalJSON(bankStateBz, bankState)
		if err != nil {
			panic(err)
		}

		stakingAddr := authtypes.NewModuleAddress(stakingtypes.NotBondedPoolName).String()
		var found bool
		for _, balance := range bankState.Balances {
			if balance.Address == stakingAddr {
				found = true
				break
			}
		}
		if !found {
			bankState.Balances = append(bankState.Balances, banktypes.Balance{
				Address: stakingAddr,
				Coins:   sdk.NewCoins(notBondedCoins),
			})
		}

		// change appState back
		rawState[stakingtypes.ModuleName] = cdc.MustMarshalJSON(stakingState)
		rawState[banktypes.ModuleName] = cdc.MustMarshalJSON(bankState)

		// replace appstate
		appState, err = json.Marshal(rawState)
		if err != nil {
			panic(err)
		}
		return appState, simAccs, chainID, genesisTimestamp
	}
}

// AppStateRandomizedFn creates calls each module's GenesisState generator function
// and creates the simulation params
func AppStateRandomizedFn(
	simManager *module.SimulationManager, r *rand.Rand, cdc codec.JSONCodec,
	accs []simtypes.Account, genesisTimestamp time.Time, appParams simtypes.AppParams,
	modules map[string]module.AppModule,
) (json.RawMessage, []simtypes.Account) {
	numAccs := int64(len(accs))

	genesisState := make(map[string]json.RawMessage)
	for name, m := range modules {
		genesisState[name] = m.DefaultGenesis(cdc)
	}

	// generate a random amount of initial stake coins and a random initial
	// number of bonded accounts
	var initialStake, numInitiallyBonded int64
	appParams.GetOrGenerate(
		cdc, simappparams.StakePerAccount, &initialStake, r,
		func(r *rand.Rand) { initialStake = r.Int63n(1e12) },
	)
	appParams.GetOrGenerate(
		cdc, simappparams.InitiallyBondedValidators, &numInitiallyBonded, r,
		func(r *rand.Rand) { numInitiallyBonded = int64(r.Intn(300)) },
	)

	if numInitiallyBonded > numAccs {
		numInitiallyBonded = numAccs
	}

	fmt.Printf(
		`Selected randomly generated parameters for simulated genesis:
{
  stake_per_account: "%d",
  initially_bonded_validators: "%d"
}
`, initialStake, numInitiallyBonded,
	)

	simState := &module.SimulationState{
		AppParams:    appParams,
		Cdc:          cdc,
		Rand:         r,
		GenState:     genesisState,
		Accounts:     accs,
		InitialStake: initialStake,
		NumBonded:    numInitiallyBonded,
		GenTimestamp: genesisTimestamp,
	}

	simManager.GenerateGenesisStates(simState)

	appState, err := json.Marshal(genesisState)
	if err != nil {
		panic(err)
	}

	return appState, accs
}

// AppStateFromGenesisFileFn util function to generate the genesis AppState
// from a genesis.json file.
func AppStateFromGenesisFileFn(r io.Reader, cdc codec.JSONCodec, genesisFile string) (tmtypes.GenesisDoc, []simtypes.Account) {
	bytes, err := ioutil.ReadFile(genesisFile)
	if err != nil {
		panic(err)
	}

	var genesis tmtypes.GenesisDoc
	// NOTE: Tendermint uses a custom JSON decoder for GenesisDoc
	err = tmjson.Unmarshal(bytes, &genesis)
	if err != nil {
		panic(err)
	}

	var appState map[string]json.RawMessage
	err = json.Unmarshal(genesis.AppState, &appState)
	if err != nil {
		panic(err)
	}

	var authGenesis authtypes.GenesisState
	if appState[authtypes.ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[authtypes.ModuleName], &authGenesis)
	}

	newAccs := make([]simtypes.Account, len(authGenesis.Accounts))
	for i, acc := range authGenesis.Accounts {
		// Pick a random private key, since we don't know the actual key
		// This should be fine as it's only used for mock Tendermint validators
		// and these keys are never actually used to sign by mock Tendermint.
		privkeySeed := make([]byte, 15)
		if _, err := r.Read(privkeySeed); err != nil {
			panic(err)
		}

		privKey := secp256k1.GenPrivKeyFromSecret(privkeySeed)

		a, ok := acc.GetCachedValue().(authtypes.AccountI)
		if !ok {
			panic("expected account")
		}

		// create simulator accounts
		simAcc := simtypes.Account{PrivKey: privKey, PubKey: privKey.PubKey(), Address: a.GetAddress()}
		newAccs[i] = simAcc
	}

	return genesis, newAccs
}
