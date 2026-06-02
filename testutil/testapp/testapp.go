package testapp

import (
	"encoding/json"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"

	coreheader "cosmossdk.io/core/header"
	"cosmossdk.io/log/v2"
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Setup creates an SDKApp initialized with one validator and one funded genesis account.
func Setup(tb testing.TB) *app.SDKApp {
	tb.Helper()

	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	if err != nil {
		tb.Fatalf("failed to get pub key: %v", err)
	}
	valSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{
		cmttypes.NewValidator(pubKey, 1),
	})

	priv := secp256k1.GenPrivKey()
	ba := authtypes.NewBaseAccount(priv.PubKey().Address().Bytes(), priv.PubKey(), 0, 0)
	genAccs := []authtypes.GenesisAccount{ba}
	balances := []banktypes.Balance{{
		Address: ba.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100000000000000))),
	}}

	return SetupWithGenesisValSet(tb, valSet, genAccs, balances...)
}

// SetupWithGenesisValSet creates an SDKApp initialized with the given validator set and genesis accounts.
func SetupWithGenesisValSet(tb testing.TB, valSet *cmttypes.ValidatorSet, genAccs []authtypes.GenesisAccount, balances ...banktypes.Balance) *app.SDKApp {
	tb.Helper()

	opts := simtestutil.AppOptionsMap{
		flags.FlagHome:    tb.TempDir(),
		flags.FlagChainID: "test-chain",
	}

	cfg := app.DefaultSDKAppConfig("app", opts)
	sdkApp := app.NewSDKApp(log.NewNopLogger(), dbm.NewMemDB(), nil, cfg)
	sdkApp.LoadModules()

	if err := sdkApp.LoadLatestVersion(); err != nil {
		tb.Fatalf("failed to load latest version: %v", err)
	}

	genesisState, err := genesisStateWithValSet(sdkApp.AppCodec(), sdkApp.DefaultGenesis(), valSet, genAccs, balances...)
	if err != nil {
		tb.Fatalf("failed to create genesis state: %v", err)
	}

	stateBytes, err := cmtjson.MarshalIndent(genesisState, "", " ")
	if err != nil {
		tb.Fatalf("failed to marshal genesis state: %v", err)
	}

	_, err = sdkApp.InitChain(&abci.RequestInitChain{
		ChainId:         "test-chain",
		Validators:      []abci.ValidatorUpdate{},
		ConsensusParams: simtestutil.DefaultConsensusParams,
		AppStateBytes:   stateBytes,
	})
	if err != nil {
		tb.Fatalf("failed to init chain: %v", err)
	}

	_, err = sdkApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             sdkApp.LastBlockHeight() + 1,
		NextValidatorsHash: valSet.Hash(),
	})
	if err != nil {
		tb.Fatalf("failed to finalize block: %v", err)
	}

	return sdkApp
}

// genesisStateWithValSet returns a new genesis state with the validator set.
func genesisStateWithValSet(
	cdc codec.Codec,
	genesisState map[string]json.RawMessage,
	valSet *cmttypes.ValidatorSet,
	genAccs []authtypes.GenesisAccount,
	balances ...banktypes.Balance,
) (map[string]json.RawMessage, error) {
	// set genesis accounts
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = cdc.MustMarshalJSON(authGenesis)

	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))

	bondAmt := sdk.DefaultPowerReduction

	for _, val := range valSet.Validators {
		pk, err := cryptocodec.FromCmtPubKeyInterface(val.PubKey)
		if err != nil {
			return nil, err
		}

		pkAny, err := codectypes.NewAnyWithValue(pk)
		if err != nil {
			return nil, err
		}

		validator := stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   sdkmath.LegacyOneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec()),
			MinSelfDelegation: sdkmath.ZeroInt(),
		}
		validators = append(validators, validator)
		delegations = append(delegations, stakingtypes.NewDelegation(genAccs[0].GetAddress().String(), sdk.ValAddress(val.Address).String(), sdkmath.LegacyOneDec()))
	}

	stakingGenesis := stakingtypes.NewGenesisState(stakingtypes.DefaultParams(), validators, delegations)
	genesisState[stakingtypes.ModuleName] = cdc.MustMarshalJSON(stakingGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		totalSupply = totalSupply.Add(b.Coins...)
	}

	for range delegations {
		totalSupply = totalSupply.Add(sdk.NewCoin(sdk.DefaultBondDenom, bondAmt))
	}

	// add bonded amount to bonded pool module account
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, bondAmt)},
	})

	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply, []banktypes.Metadata{}, []banktypes.SendEnabled{})
	genesisState[banktypes.ModuleName] = cdc.MustMarshalJSON(bankGenesis)

	return genesisState, nil
}

// NewContext returns a new sdk.Context backed by the finalizeBlock MultiStore,
// with the chain ID set from the app configuration. This is the preferred way
// to get a test context instead of calling sdkApp.NewContext(false) directly,
// which returns a context with an empty chain ID.
func NewContext(sdkApp *app.SDKApp) sdk.Context {
	return sdkApp.NewContext(false).WithChainID(sdkApp.ChainID())
}

// NextBlock advances the app by one block. It finalizes the current block,
// commits it, and returns a new context for the next block at blockTime+jumpTime.
func NextBlock(sdkApp *app.SDKApp, ctx sdk.Context, jumpTime time.Duration) (sdk.Context, error) {
	_, err := sdkApp.FinalizeBlock(&abci.RequestFinalizeBlock{Height: ctx.BlockHeight(), Time: ctx.BlockTime()})
	if err != nil {
		return sdk.Context{}, err
	}
	_, err = sdkApp.Commit()
	if err != nil {
		return sdk.Context{}, err
	}

	newBlockTime := ctx.BlockTime().Add(jumpTime)

	header := cmtproto.Header{
		Height:  ctx.BlockHeight() + 1,
		Time:    newBlockTime,
		ChainID: ctx.ChainID(),
	}

	newCtx := sdkApp.BaseApp.NewNextBlockContext(header).WithHeaderInfo(coreheader.Info{
		Height: header.Height,
		Time:   header.Time,
	})

	return newCtx, nil
}
