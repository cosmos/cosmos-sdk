// this file is a port of testutil/sims/app_helpers.go from v1 to v2 architecture
package integration

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsign "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
	"time"

	"cosmossdk.io/core/comet"
	corecontext "cosmossdk.io/core/context"
	"cosmossdk.io/core/server"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/runtime/v2"
	"cosmossdk.io/runtime/v2/services"
	"cosmossdk.io/server/v2/stf"
	"cosmossdk.io/server/v2/stf/branch"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	consensustypes "cosmossdk.io/x/consensus/types"
	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmttypes "github.com/cometbft/cometbft/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

const DefaultGenTxGas = 10000000

type stateMachineTx = transaction.Tx

// DefaultConsensusParams defines the default CometBFT consensus params used in
// SimApp testing.
var DefaultConsensusParams = &cmtproto.ConsensusParams{
	Version: &cmtproto.VersionParams{
		App: 1,
	},
	Block: &cmtproto.BlockParams{
		MaxBytes: 200000,
		MaxGas:   100_000_000,
	},
	Evidence: &cmtproto.EvidenceParams{
		MaxAgeNumBlocks: 302400,
		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
		MaxBytes:        10000,
	},
	Validator: &cmtproto.ValidatorParams{
		PubKeyTypes: []string{
			cmttypes.ABCIPubKeyTypeEd25519,
			cmttypes.ABCIPubKeyTypeSecp256k1,
		},
	},
}

// StartupConfig defines the startup configuration new a test application.
//
// ValidatorSet defines a custom validator set to be validating the app.
// BaseAppOption defines the additional operations that must be run on baseapp before app start.
// AtGenesis defines if the app started should already have produced block or not.
type StartupConfig struct {
	ValidatorSet    func() (*cmttypes.ValidatorSet, error)
	AppOption       runtime.AppBuilderOption[stateMachineTx]
	AtGenesis       bool
	GenesisAccounts []GenesisAccount
	HomeDir         string
}

func DefaultStartUpConfig() StartupConfig {
	priv := secp256k1.GenPrivKey()
	ba := authtypes.NewBaseAccount(
		priv.PubKey().Address().Bytes(),
		priv.PubKey(),
		0,
		0,
	)
	ga := GenesisAccount{
		ba,
		sdk.NewCoins(
			sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100000000000000)),
		),
	}
	return StartupConfig{
		ValidatorSet:    CreateRandomValidatorSet,
		AtGenesis:       false,
		GenesisAccounts: []GenesisAccount{ga},
	}
}

// Setup initializes a new runtime.App and can inject values into extraOutputs.
// It uses SetupWithConfiguration under the hood.
func Setup(
	appConfig depinject.Config,
	extraOutputs ...interface{},
) (*App, error) {
	return SetupWithConfiguration(
		appConfig,
		DefaultStartUpConfig(),
		extraOutputs...)
}

// SetupAtGenesis initializes a new runtime.App at genesis and can inject values into extraOutputs.
// It uses SetupWithConfiguration under the hood.
func SetupAtGenesis(
	appConfig depinject.Config,
	extraOutputs ...interface{},
) (*App, error) {
	cfg := DefaultStartUpConfig()
	cfg.AtGenesis = true
	return SetupWithConfiguration(appConfig, cfg, extraOutputs...)
}

// SetupWithConfiguration initializes a new runtime.App. A Nop logger is set in runtime.App.
// appConfig defines the application configuration (f.e. app_config.go).
// extraOutputs defines the extra outputs to be assigned by the dependency injector (depinject).
func SetupWithConfiguration(
	appConfig depinject.Config,
	startupConfig StartupConfig,
	extraOutputs ...interface{},
) (*App, error) {
	// create the app with depinject
	var (
		app             *runtime.App[stateMachineTx]
		appBuilder      *runtime.AppBuilder[stateMachineTx]
		storeBuilder    *runtime.StoreBuilder
		txConfig        client.TxConfig
		txConfigOptions tx.ConfigOptions
		cometService    comet.Service                   = &cometServiceImpl{}
		kvFactory       corestore.KVStoreServiceFactory = func(actor []byte) corestore.KVStoreService {
			return services.NewGenesisKVService(actor, &storeService{actor, stf.NewKVStoreService(actor)})
		}
		cdc codec.Codec
		err error
	)

	if err := depinject.Inject(
		depinject.Configs(
			appConfig,
			codec.DefaultProviders,
			depinject.Supply(
				services.NewGenesisHeaderService(stf.HeaderService{}),
				&dynamicConfigImpl{startupConfig.HomeDir},
				cometService,
				kvFactory,
				&eventService{},
			),
			depinject.Invoke(
				std.RegisterInterfaces,
			),
		),
		append(extraOutputs, &appBuilder, &cdc, &txConfigOptions, &txConfig, &storeBuilder)...); err != nil {
		return nil, fmt.Errorf("failed to inject dependencies: %w", err)
	}

	app, err = appBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build app: %w", err)
	}
	if err := app.LoadLatest(); err != nil {
		return nil, fmt.Errorf("failed to load app: %w", err)
	}

	// create validator set
	valSet, err := startupConfig.ValidatorSet()
	if err != nil {
		return nil, errors.New("failed to create validator set")
	}

	var (
		balances    []banktypes.Balance
		genAccounts []authtypes.GenesisAccount
	)
	for _, ga := range startupConfig.GenesisAccounts {
		genAccounts = append(genAccounts, ga.GenesisAccount)
		balances = append(
			balances,
			banktypes.Balance{
				Address: ga.GenesisAccount.GetAddress().String(),
				Coins:   ga.Coins,
			},
		)
	}

	genesisJSON, err := genesisStateWithValSet(
		cdc,
		app.DefaultGenesis(),
		valSet,
		genAccounts,
		balances...)
	if err != nil {
		return nil, fmt.Errorf("failed to create genesis state: %w", err)
	}

	// init chain must be called to stop deliverState from being nil
	genesisJSONBytes, err := cmtjson.MarshalIndent(genesisJSON, "", " ")
	if err != nil {
		return nil, fmt.Errorf(
			"failed to marshal default genesis state: %w",
			err,
		)
	}

	ctx := context.WithValue(
		context.Background(),
		corecontext.CometParamsInitInfoKey,
		&consensustypes.MsgUpdateParams{
			Authority: "consensus",
			Block:     DefaultConsensusParams.Block,
			Evidence:  DefaultConsensusParams.Evidence,
			Validator: DefaultConsensusParams.Validator,
			Abci:      DefaultConsensusParams.Abci,
			Synchrony: DefaultConsensusParams.Synchrony,
			Feature:   DefaultConsensusParams.Feature,
		},
	)

	store := storeBuilder.Get()
	if store == nil {
		return nil, fmt.Errorf("failed to build store: %w", err)
	}
	err = store.SetInitialVersion(1)
	if err != nil {
		return nil, fmt.Errorf("failed to set initial version: %w", err)
	}
	integrationApp := &App{App: app, Store: store, txConfig: txConfig}

	emptyHash := sha256.Sum256(nil)
	_, genesisState, err := app.InitGenesis(
		ctx,
		&server.BlockRequest[stateMachineTx]{
			Height:    1,
			Time:      time.Now(),
			Hash:      emptyHash[:],
			ChainId:   "test-chain",
			AppHash:   emptyHash[:],
			IsGenesis: true,
		},
		genesisJSONBytes,
		&genesisTxCodec{txConfigOptions},
	)
	if err != nil {
		return nil, fmt.Errorf("failed init genesiss: %w", err)
	}

	_, err = integrationApp.Commit(genesisState)
	if err != nil {
		return nil, fmt.Errorf("failed to commit initial version: %w", err)
	}

	return integrationApp, nil
}

type App struct {
	*runtime.App[stateMachineTx]
	Store    runtime.Store
	txConfig client.TxConfig
}

func (a *App) Run(
	ctx context.Context,
	state corestore.ReaderMap,
	fn func(ctx context.Context) error,
) (corestore.ReaderMap, error) {
	nextState := branch.DefaultNewWriterMap(state)
	iCtx := integrationContext{state: nextState}
	ctx = context.WithValue(ctx, contextKey, iCtx)
	err := fn(ctx)
	if err != nil {
		return nil, err
	}
	return nextState, nil
}

func (a *App) Commit(state corestore.WriterMap) ([]byte, error) {
	changes, err := state.GetStateChanges()
	if err != nil {
		return nil, fmt.Errorf("failed to get state changes: %w", err)
	}
	cs := &corestore.Changeset{Changes: changes}
	//for _, change := range changes {
	//	if !bytes.Equal(change.Actor, []byte("acc")) {
	//		continue
	//	}
	//	for _, kv := range change.StateChanges {
	//		fmt.Printf("actor: %s, key: %x, value: %x\n", change.Actor, kv.Key, kv.Value)
	//	}
	//}
	return a.Store.Commit(cs)
}

func (a *App) SignCheckDeliver(
	t *testing.T, ctx context.Context, height uint64, msgs []sdk.Msg,
	chainID string, accNums, accSeqs []uint64, privateKeys []cryptotypes.PrivKey,
	txErrString string,
) server.TxResult {
	t.Helper()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	sigs := make([]signing.SignatureV2, len(privateKeys))

	// create a random length memo
	memo := simulation.RandStringOfLength(r, simulation.RandIntBetween(r, 0, 100))

	signMode, err := authsign.APISignModeToInternal(a.txConfig.SignModeHandler().DefaultMode())
	require.NoError(t, err)

	// 1st round: set SignatureV2 with empty signatures, to set correct
	// signer infos.
	for i, p := range privateKeys {
		sigs[i] = signing.SignatureV2{
			PubKey: p.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode: signMode,
			},
			Sequence: accSeqs[i],
		}
	}

	txBuilder := a.txConfig.NewTxBuilder()
	err = txBuilder.SetMsgs(msgs...)
	require.NoError(t, err)
	err = txBuilder.SetSignatures(sigs...)
	require.NoError(t, err)
	txBuilder.SetMemo(memo)
	txBuilder.SetFeeAmount(sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)})
	txBuilder.SetGasLimit(DefaultGenTxGas)

	// 2nd round: once all signer infos are set, every signer can sign.
	for i, p := range privateKeys {
		signerData := authsign.SignerData{
			Address:       sdk.AccAddress(p.PubKey().Address()).String(),
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
			PubKey:        p.PubKey(),
		}

		signBytes, err := authsign.GetSignBytesAdapter(
			ctx, a.txConfig.SignModeHandler(), signMode, signerData,
			// todo why fetch twice?
			txBuilder.GetTx())
		require.NoError(t, err)
		sig, err := p.Sign(signBytes)
		require.NoError(t, err)
		sigs[i].Data.(*signing.SingleSignatureData).Signature = sig
	}
	err = txBuilder.SetSignatures(sigs...)
	require.NoError(t, err)

	builtTx := txBuilder.GetTx()
	require.NoError(t, err)

	blockResponse, blockState, err := a.DeliverBlock(ctx, &server.BlockRequest[stateMachineTx]{
		Height:  height,
		Txs:     []stateMachineTx{builtTx},
		Hash:    make([]byte, 32),
		AppHash: make([]byte, 32),
	})
	require.NoError(t, err)

	require.Equal(t, 1, len(blockResponse.TxResults))
	txResult := blockResponse.TxResults[0]
	finalizeSuccess := txResult.Code == 0
	if txErrString != "" {
		require.False(t, finalizeSuccess)
		require.ErrorContains(t, txResult.Error, txErrString)
	} else {
		require.True(t, finalizeSuccess)
		require.NoError(t, txResult.Error)
	}

	_, err = a.Commit(blockState)
	require.NoError(t, err)

	return txResult
}

func (a *App) CheckBalance(
	ctx context.Context, t *testing.T, addr sdk.AccAddress, expected sdk.Coins, keeper bankkeeper.Keeper,
) {
	t.Helper()
	balances := keeper.GetAllBalances(ctx, addr)
	require.Equal(t, expected, balances)
}
