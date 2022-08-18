package sims

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"cosmossdk.io/depinject"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// SimAppChainID hardcoded chainID for simulation
const (
	DefaultGenTxGas = 10000000
	SimAppChainID   = "simulation-app"
)

// DefaultConsensusParams defines the default Tendermint consensus params used in
// SimApp testing.
var DefaultConsensusParams = &tmproto.ConsensusParams{
	Block: &tmproto.BlockParams{
		MaxBytes: 200000,
		MaxGas:   2000000,
	},
	Evidence: &tmproto.EvidenceParams{
		MaxAgeNumBlocks: 302400,
		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
		MaxBytes:        10000,
	},
	Validator: &tmproto.ValidatorParams{
		PubKeyTypes: []string{
			tmtypes.ABCIPubKeyTypeEd25519,
		},
	},
}

// CreateRandomValidatorSet creates a validator set with one random validator
func CreateRandomValidatorSet() (*tmtypes.ValidatorSet, error) {
	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to get pub key: %w", err)
	}

	// create validator set with single validator
	validator := tmtypes.NewValidator(pubKey, 1)

	return tmtypes.NewValidatorSet([]*tmtypes.Validator{validator}), nil
}

// Setup initializes a new runtime.App and can inject values into extraOutputs.
// It uses SetupWithConfiguration under the hood.
func Setup(appConfig depinject.Config, extraOutputs ...interface{}) (*runtime.App, error) {
	return SetupWithConfiguration(appConfig, CreateRandomValidatorSet, nil, false, extraOutputs...)
}

// SetupAtGenesis initializes a new runtime.App at genesis and can inject values into extraOutputs.
// It uses SetupWithConfiguration under the hood.
func SetupAtGenesis(appConfig depinject.Config, extraOutputs ...interface{}) (*runtime.App, error) {
	return SetupWithConfiguration(appConfig, CreateRandomValidatorSet, nil, true, extraOutputs...)
}

// SetupWithBaseAppOption initializes a new runtime.App and can inject values into extraOutputs.
// With specific baseApp options. It uses SetupWithConfiguration under the hood.
func SetupWithBaseAppOption(appConfig depinject.Config, baseAppOption runtime.BaseAppOption, extraOutputs ...interface{}) (*runtime.App, error) {
	return SetupWithConfiguration(appConfig, CreateRandomValidatorSet, baseAppOption, false, extraOutputs...)
}

// SetupWithConfiguration initializes a new runtime.App. A Nop logger is set in runtime.App.
// appConfig usually load from a `app.yaml` with `appconfig.LoadYAML`, defines the application configuration.
// validatorSet defines a custom validator set to be validating the app.
// baseAppOption defines the additional operations that must be run on baseapp before app start.
// genesis defines if the app started should already have produced block or not.
// extraOutputs defines the extra outputs to be assigned by the dependency injector (depinject).
func SetupWithConfiguration(appConfig depinject.Config, validatorSet func() (*tmtypes.ValidatorSet, error), baseAppOption runtime.BaseAppOption, genesis bool, extraOutputs ...interface{}) (*runtime.App, error) {
	// create the app with depinject
	var (
		app        *runtime.App
		appBuilder *runtime.AppBuilder
		codec      codec.Codec
	)

	if err := depinject.Inject(
		appConfig,
		append(extraOutputs, &appBuilder, &codec)...,
	); err != nil {
		return nil, fmt.Errorf("failed to inject dependencies: %w", err)
	}

	if baseAppOption != nil {
		app = appBuilder.Build(log.NewNopLogger(), dbm.NewMemDB(), nil, baseAppOption)
	} else {
		app = appBuilder.Build(log.NewNopLogger(), dbm.NewMemDB(), nil)
	}
	if err := app.Load(true); err != nil {
		return nil, fmt.Errorf("failed to load app: %w", err)
	}

	// create validator set
	valSet, err := validatorSet()
	if err != nil {
		return nil, fmt.Errorf("failed to create validator set")
	}

	// generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000000000000))),
	}

	genesisState, err := GenesisStateWithValSet(codec, appBuilder.DefaultGenesis(), valSet, []authtypes.GenesisAccount{acc}, balance)
	if err != nil {
		return nil, fmt.Errorf("failed to create genesis state: %w", err)
	}

	// init chain must be called to stop deliverState from being nil
	stateBytes, err := tmjson.MarshalIndent(genesisState, "", " ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal default genesis state: %w", err)
	}

	// init chain will set the validator set and initialize the genesis accounts
	app.InitChain(
		abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)

	// commit genesis changes
	if !genesis {
		app.Commit()
		app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{
			Height:             app.LastBlockHeight() + 1,
			AppHash:            app.LastCommitID().Hash,
			ValidatorsHash:     valSet.Hash(),
			NextValidatorsHash: valSet.Hash(),
		}})
	}

	return app, nil
}

// GenesisStateWithValSet returns a new genesis state with the validator set
func GenesisStateWithValSet(codec codec.Codec, genesisState map[string]json.RawMessage,
	valSet *tmtypes.ValidatorSet, genAccs []authtypes.GenesisAccount,
	balances ...banktypes.Balance,
) (map[string]json.RawMessage, error) {
	// set genesis accounts
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = codec.MustMarshalJSON(authGenesis)

	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))

	bondAmt := sdk.DefaultPowerReduction

	for _, val := range valSet.Validators {
		pk, err := cryptocodec.FromTmPubKeyInterface(val.PubKey)
		if err != nil {
			return nil, fmt.Errorf("failed to convert pubkey: %w", err)
		}

		pkAny, err := codectypes.NewAnyWithValue(pk)
		if err != nil {
			return nil, fmt.Errorf("failed to create new any: %w", err)
		}

		validator := stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   math.LegacyOneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec()),
			MinSelfDelegation: math.ZeroInt(),
		}
		validators = append(validators, validator)
		delegations = append(delegations, stakingtypes.NewDelegation(genAccs[0].GetAddress(), val.Address.Bytes(), math.LegacyOneDec()))

	}
	// set validators and delegations
	stakingGenesis := stakingtypes.NewGenesisState(stakingtypes.DefaultParams(), validators, delegations)
	genesisState[stakingtypes.ModuleName] = codec.MustMarshalJSON(stakingGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens to total supply
		totalSupply = totalSupply.Add(b.Coins...)
	}

	for range delegations {
		// add delegated tokens to total supply
		totalSupply = totalSupply.Add(sdk.NewCoin(sdk.DefaultBondDenom, bondAmt))
	}

	// add bonded amount to bonded pool module account
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, bondAmt)},
	})

	// update total supply
	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply, []banktypes.Metadata{}, []banktypes.SendEnabled{})
	genesisState[banktypes.ModuleName] = codec.MustMarshalJSON(bankGenesis)

	return genesisState, nil
}

// EmptyAppOptions is a stub implementing AppOptions
type EmptyAppOptions struct{}

// Get implements AppOptions
func (ao EmptyAppOptions) Get(o string) interface{} {
	return nil
}

// AppOptionsMap is a stub implementing AppOptions which can get data from a map
type AppOptionsMap map[string]interface{}

func (m AppOptionsMap) Get(key string) interface{} {
	v, ok := m[key]
	if !ok {
		return interface{}(nil)
	}

	return v
}

func NewAppOptionsWithFlagHome(homePath string) servertypes.AppOptions {
	return AppOptionsMap{
		flags.FlagHome: homePath,
	}
}
