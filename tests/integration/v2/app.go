package integration

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"

	corebranch "cosmossdk.io/core/branch"
	"cosmossdk.io/core/comet"
	corecontext "cosmossdk.io/core/context"
	"cosmossdk.io/core/gas"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/server"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/depinject"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/runtime/v2"
	"cosmossdk.io/runtime/v2/services"
	"cosmossdk.io/server/v2/stf"
	"cosmossdk.io/server/v2/stf/branch"
	"cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/root"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	consensustypes "cosmossdk.io/x/consensus/types"
	txsigning "cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsign "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

const DefaultGenTxGas = 10000000
const (
	Genesis_COMMIT = iota
	Genesis_NOCOMMIT
	Genesis_SKIP
)

type stateMachineTx = transaction.Tx

type handler = func(ctx context.Context) (transaction.Msg, error)

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

// StartupConfig defines the startup configuration of a new test app.
type StartupConfig struct {
	// ValidatorSet defines a custom validator set to be validating the app.
	ValidatorSet func() (*cmttypes.ValidatorSet, error)
	// AppOption defines the additional operations that will be run in the app builder phase.
	AppOption runtime.AppBuilderOption[stateMachineTx]
	// GenesisBehavior defines the behavior of the app at genesis.
	GenesisBehavior int
	// GenesisAccounts defines the genesis accounts to be used in the app.
	GenesisAccounts []GenesisAccount
	// HomeDir defines the home directory of the app where config and data will be stored.
	HomeDir string
	// BranchService defines the custom branch service to be used in the app.
	BranchService corebranch.Service
	// RouterServiceBuilder defines the custom builder
	// for msg router and query router service to be used in the app.
	RouterServiceBuilder runtime.RouterServiceBuilder
	// HeaderService defines the custom header service to be used in the app.
	HeaderService header.Service

	GasService gas.Service
}

func DefaultStartUpConfig(tb testing.TB) StartupConfig {
	tb.Helper()

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
	homedir := tb.TempDir()
	tb.Logf("generated integration test app config; HomeDir=%s", homedir)
	return StartupConfig{
		ValidatorSet:    CreateRandomValidatorSet,
		GenesisBehavior: Genesis_COMMIT,
		GenesisAccounts: []GenesisAccount{ga},
		HomeDir:         homedir,
		BranchService:   stf.BranchService{},
		RouterServiceBuilder: runtime.NewRouterBuilder(
			stf.NewMsgRouterService, stf.NewQueryRouterService(),
		),
		HeaderService: services.NewGenesisHeaderService(stf.HeaderService{}),
		GasService:    stf.NewGasMeterService(),
	}
}

// RunMsgConfig defines the run message configuration.
type RunMsgConfig struct {
	Commit bool
}

// Option is a function that can be used to configure the integration app.
type Option func(*RunMsgConfig)

// WithAutomaticCommit enables automatic commit.
// This means that the integration app will automatically commit the state after each msg.
func WithAutomaticCommit() Option {
	return func(cfg *RunMsgConfig) {
		cfg.Commit = true
	}
}

// NewApp initializes a new runtime.App. A Nop logger is set in runtime.App.
// appConfig defines the application configuration (f.e. app_config.go).
// extraOutputs defines the extra outputs to be assigned by the dependency injector (depinject).
func NewApp(
	appConfig depinject.Config,
	startupConfig StartupConfig,
	extraOutputs ...interface{},
) (*App, error) {
	// create the app with depinject
	var (
		storeBuilder    = root.NewBuilder()
		app             *runtime.App[stateMachineTx]
		appBuilder      *runtime.AppBuilder[stateMachineTx]
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
				&root.Config{
					Home:         startupConfig.HomeDir,
					AppDBBackend: "goleveldb",
					Options:      root.DefaultStoreOptions(),
				},
				runtime.GlobalConfig{
					"server": server.ConfigMap{
						"minimum-gas-prices": "0stake",
					},
				},
				cometService,
				kvFactory,
				&eventService{},
				storeBuilder,
				startupConfig.BranchService,
				startupConfig.RouterServiceBuilder,
				startupConfig.HeaderService,
				startupConfig.GasService,
			),
			depinject.Invoke(
				std.RegisterInterfaces,
				std.RegisterLegacyAminoCodec,
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

	store := storeBuilder.Get()
	if store == nil {
		return nil, fmt.Errorf("failed to build store: %w", err)
	}
	err = store.SetInitialVersion(0)
	if err != nil {
		return nil, fmt.Errorf("failed to set initial version: %w", err)
	}

	integrationApp := &App{App: app, Store: store, txConfig: txConfig, lastHeight: 0}
	if startupConfig.GenesisBehavior == Genesis_SKIP {
		return integrationApp, nil
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
		return nil, fmt.Errorf("failed init genesis: %w", err)
	}

	if startupConfig.GenesisBehavior == Genesis_NOCOMMIT {
		integrationApp.lastHeight = 0
		return integrationApp, nil
	}

	_, err = integrationApp.Commit(genesisState)
	if err != nil {
		return nil, fmt.Errorf("failed to commit initial version: %w", err)
	}

	return integrationApp, nil
}

// App is a wrapper around runtime.App that provides additional testing utilities.
type App struct {
	*runtime.App[stateMachineTx]
	lastHeight uint64
	Store      store.RootStore
	txConfig   client.TxConfig
}

func (a App) LastBlockHeight() uint64 {
	return a.lastHeight
}

// Deliver delivers a block with the given transactions and returns the resulting state.
func (a *App) Deliver(
	t *testing.T, ctx context.Context, txs []stateMachineTx,
) (*server.BlockResponse, corestore.WriterMap) {
	t.Helper()
	req := &server.BlockRequest[stateMachineTx]{
		Height:  a.lastHeight + 1,
		Txs:     txs,
		Hash:    make([]byte, 32),
		AppHash: make([]byte, 32),
	}
	resp, state, err := a.DeliverBlock(ctx, req)
	require.NoError(t, err)
	a.lastHeight++

	// update block height and block time if integration context is present
	iCtx, ok := ctx.Value(contextKey).(*integrationContext)
	if ok {
		iCtx.header.Height = int64(a.lastHeight)
	}
	return resp, state
}

// StateLatestContext creates returns a new context from context.Background() with the latest state.
func (a *App) StateLatestContext(tb testing.TB) context.Context {
	tb.Helper()
	_, state, err := a.Store.StateLatest()
	require.NoError(tb, err)
	writeableState := branch.DefaultNewWriterMap(state)
	iCtx := &integrationContext{state: writeableState}
	return context.WithValue(context.Background(), contextKey, iCtx)
}

// Commit commits the given state and returns the new state hash.
func (a *App) Commit(state corestore.WriterMap) ([]byte, error) {
	changes, err := state.GetStateChanges()
	if err != nil {
		return nil, fmt.Errorf("failed to get state changes: %w", err)
	}
	cs := &corestore.Changeset{Version: a.lastHeight, Changes: changes}
	return a.Store.Commit(cs)
}

// SignCheckDeliver signs and checks the given messages and delivers them.
func (a *App) SignCheckDeliver(
	t *testing.T, ctx context.Context, msgs []sdk.Msg,
	chainID string, accNums, accSeqs []uint64, privateKeys []cryptotypes.PrivKey,
	txErrString string,
) server.TxResult {
	t.Helper()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	sigs := make([]signing.SignatureV2, len(privateKeys))

	// create a random length memo
	memo := simulation.RandStringOfLength(r, simulation.RandIntBetween(r, 0, 100))

	// 1st round: set SignatureV2 with empty signatures, to set correct
	// signer infos.
	for i, p := range privateKeys {
		sigs[i] = signing.SignatureV2{
			PubKey: p.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode: a.txConfig.SignModeHandler().DefaultMode(),
			},
			Sequence: accSeqs[i],
		}
	}

	txBuilder := a.txConfig.NewTxBuilder()
	err := txBuilder.SetMsgs(msgs...)
	require.NoError(t, err)
	err = txBuilder.SetSignatures(sigs...)
	require.NoError(t, err)
	txBuilder.SetMemo(memo)
	txBuilder.SetFeeAmount(sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)})
	txBuilder.SetGasLimit(DefaultGenTxGas)

	// 2nd round: once all signer infos are set, every signer can sign.
	for i, p := range privateKeys {
		anyPk, err := codectypes.NewAnyWithValue(p.PubKey())
		require.NoError(t, err)

		signerData := txsigning.SignerData{
			Address:       sdk.AccAddress(p.PubKey().Address()).String(),
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
			PubKey:        &anypb.Any{TypeUrl: anyPk.TypeUrl, Value: anyPk.Value},
		}

		signBytes, err := authsign.GetSignBytesAdapter(
			ctx, a.txConfig.SignModeHandler(), a.txConfig.SignModeHandler().DefaultMode(), signerData,
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
	blockResponse, blockState := a.Deliver(t, ctx, []stateMachineTx{builtTx})

	require.Equal(t, 1, len(blockResponse.TxResults))
	txResult := blockResponse.TxResults[0]
	if txErrString != "" {
		require.ErrorContains(t, txResult.Error, txErrString)
	} else {
		require.NoError(t, txResult.Error)
	}

	_, err = a.Commit(blockState)
	require.NoError(t, err)

	return txResult
}

// RunMsg runs the handler for a transaction message.
// It required the context to have the integration context.
// a new state is committed if the option WithAutomaticCommit is set in options.
func (app *App) RunMsg(t *testing.T, ctx context.Context, handler handler, option ...Option) (resp transaction.Msg, err error) {
	t.Helper()

	// set options
	cfg := &RunMsgConfig{}
	for _, opt := range option {
		opt(cfg)
	}

	// need to have integration context
	integrationCtx, ok := ctx.Value(contextKey).(*integrationContext)
	require.True(t, ok)

	resp, err = handler(ctx)

	if cfg.Commit {
		app.lastHeight++
		_, err := app.Commit(integrationCtx.state)
		require.NoError(t, err)
	}

	return resp, err
}

// CheckBalance checks the balance of the given address.
func (a *App) CheckBalance(
	t *testing.T, ctx context.Context, addr sdk.AccAddress, expected sdk.Coins, keeper bankkeeper.Keeper,
) {
	t.Helper()
	balances := keeper.GetAllBalances(ctx, addr)
	require.Equal(t, expected, balances)
}

func (a *App) Close() error {
	return a.Store.Close()
}
