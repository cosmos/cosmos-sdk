package simapp

import (
	"context"
	appmanager "cosmossdk.io/core/app"
	"cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/comet"
	corecontext "cosmossdk.io/core/context"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
	cometbfttypes "cosmossdk.io/server/v2/cometbft/types"
	banktypes "cosmossdk.io/x/bank/types"
	"encoding/json"
	"errors"
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simsx"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	"github.com/stretchr/testify/require"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type T = transaction.Tx
type HasWeightedOperationsX interface {
	WeightedOperationsX(weight simsx.WeightSource, reg simsx.Registry)
}

func TestSimsAppV2(t *testing.T) {
	DefaultNodeHome = t.TempDir()
	currentDir, err := os.Getwd()
	require.NoError(t, err)
	configPath := filepath.Join(currentDir, "testdata")
	v, err := serverv2.ReadConfig(configPath)
	require.NoError(t, err)

	logger := log.NewLogger(os.Stdout)
	app := NewSimApp[T](logger, v)
	//var abciApp *cometbft.Consensus[T]

	tCfg := cli.NewConfigFromFlags().With(t, 1, nil)

	appStateFn := simtestutil.AppStateFnY(
		app.AppCodec(),
		app.AuthKeeper.AddressCodec(),
		app.StakingKeeper.ValidatorAddressCodec(),
		toSimsModule(app.ModuleManager().Modules()),
		app.DefaultGenesis(),
	)

	r := rand.New(rand.NewSource(tCfg.Seed))
	params := simulation.RandomParams(r)
	accounts := simtypes.RandomAccounts(r, params.NumKeys())

	appState, accounts, chainID, genesisTimestamp := appStateFn(r, accounts, tCfg)

	appStore := app.GetStore().(cometbfttypes.Store)
	//consensusParams := simulation.RandomConsensusParams(r, appState, cdc, blockMaxGas)
	req := &appmanager.BlockRequest[T]{
		Height:    0,
		Time:      genesisTimestamp,
		Hash:      make([]byte, 32),
		ChainId:   chainID,
		AppHash:   make([]byte, 32),
		IsGenesis: true,
	}
	ctx, done := context.WithCancel(context.Background())
	_, genesisState, err := app.InitGenesis(ctx, req, appState, &genericTxDecoder[T]{txConfig: app.TxConfig()})
	require.NoError(t, err)
	changeSet, err := genesisState.GetStateChanges()
	require.NoError(t, err)
	cs := &store.Changeset{Changes: changeSet}
	stateRoot, err := appStore.Commit(cs)
	require.NoError(t, err)

	// next add a block
	emptySimParams := make(map[string]json.RawMessage, 0) // todo
	weights := simsx.ParamWeightSource(emptySimParams)
	reporter := simsx.NewBasicSimulationReporter(simsx.SkipHookFn(func(args ...any) { done() }))

	oReg := make(SimsV2Reg) // reporter, app.AuthKeeper, app.BankKeeper, app.txConfig.SigningContext().AddressCodec(), logger)

	// register all msg fectories // todo: just 1 example here
	w, ok := app.ModuleManager().Modules()["bank"].(HasWeightedOperationsX)
	if !ok {
		panic("alex")
	}
	w.WeightedOperationsX(weights, oReg)
	reqN := &appmanager.BlockRequest[T]{
		Height:  2,
		Time:    genesisTimestamp.Add(time.Second),
		Hash:    stateRoot,
		ChainId: chainID,
		AppHash: stateRoot,
	}
	cometInfo := comet.Info{
		//Evidence:        toCoreEvidence(req.Misbehavior),
		//ValidatorsHash:  req.NextValidatorsHash,
		//ProposerAddress: req.ProposerAddress,
		//LastCommit:      toCoreCommitInfo(req.DecidedLastCommit),
	}
	ctx = context.WithValue(ctx, corecontext.CometInfoKey, cometInfo)
	blockRsp, updates, err := app.DeliverSims(ctx, reqN, func(ctx context.Context) (T, bool) {
		// todo: sort and pick msg factory by weight
		wFac := oReg[sdk.MsgTypeURL(&banktypes.MsgSend{})]
		sRep := reporter.WithScope(wFac.factory.MsgType())

		// the stf context is required
		testData := simsx.NewChainDataSource(ctx, r, app.AuthKeeper, app.BankKeeper, app.txConfig.SigningContext().AddressCodec(), accounts...)

		signers, msg := wFac.factory.Create()(ctx, testData, sRep)
		if sRep.IsSkipped() {
			panic("alex: skipped: " + reporter.Comment()) // todo: skip and continue in loop
		}
		tx, err := genTestTX(ctx, app.AuthKeeper, signers, msg, r, app.txConfig, chainID)
		require.NoError(t, err)
		return tx, false // todo: do loop
	})
	require.NoError(t, err)
	changeSet, err = updates.GetStateChanges()
	require.NoError(t, err)
	cs = &store.Changeset{Changes: changeSet}
	stateRoot, err = appStore.Commit(cs)
	require.NoError(t, err)
	for _, v := range blockRsp.TxResults {
		require.NoError(t, v.Error)
	}
}

const defaultGas = 500_000 // todo: pick value
func genTestTX(
	ctx context.Context,
	ak simsx.AccountSource,
	senders []simsx.SimAccount,
	msg sdk.Msg,
	r *rand.Rand,
	txGen client.TxConfig,
	chainID string,
) (sdk.Tx, error) {
	accountNumbers := make([]uint64, len(senders))
	sequenceNumbers := make([]uint64, len(senders))
	for i := 0; i < len(senders); i++ {
		acc := ak.GetAccount(ctx, senders[i].Address)
		accountNumbers[i] = acc.GetAccountNumber()
		sequenceNumbers[i] = acc.GetSequence()
	}
	fees := senders[0].LiquidBalance().RandFees()
	return sims.GenSignedMockTx(
		r,
		txGen,
		[]sdk.Msg{msg},
		fees,
		sims.DefaultGenTxGas,
		chainID,
		accountNumbers,
		sequenceNumbers,
		Collect(senders, func(a simsx.SimAccount) cryptotypes.PrivKey { return a.PrivKey })...,
	)
}

type weightedFactory struct {
	weight  uint32
	factory simsx.SimMsgFactoryX
}

var _ simsx.Registry = &SimsV2Reg{}

type SimsV2Reg map[string]weightedFactory

func (s SimsV2Reg) Add(weight uint32, f simsx.SimMsgFactoryX) {
	s[sdk.MsgTypeURL(f.MsgType())] = weightedFactory{weight: weight, factory: f}
}

func toSimsModule(modules map[string]appmodule.AppModule) []module.AppModuleSimulation {
	r := make([]module.AppModuleSimulation, 0, len(modules))
	for _, v := range modules {
		if m, ok := v.(module.AppModuleSimulation); ok {
			r = append(r, m)
		}
	}
	return r
}

func mapsInsert[K comparable, V any](src []K, f func(K) V) map[K]V {
	r := make(map[K]V, len(src))
	for _, addr := range src {
		r[addr] = f(addr)
	}
	return r
}

var _ transaction.Codec[transaction.Tx] = &genericTxDecoder[transaction.Tx]{}

// todo: same as in commands
type genericTxDecoder[T transaction.Tx] struct {
	txConfig client.TxConfig
}

// Decode implements transaction.Codec.
func (t *genericTxDecoder[T]) Decode(bz []byte) (T, error) {
	var out T
	tx, err := t.txConfig.TxDecoder()(bz)
	if err != nil {
		return out, err
	}

	var ok bool
	out, ok = tx.(T)
	if !ok {
		return out, errors.New("unexpected Tx type")
	}

	return out, nil
}

// DecodeJSON implements transaction.Codec.
func (t *genericTxDecoder[T]) DecodeJSON(bz []byte) (T, error) {
	var out T
	tx, err := t.txConfig.TxJSONDecoder()(bz)
	if err != nil {
		return out, err
	}

	var ok bool
	out, ok = tx.(T)
	if !ok {
		return out, errors.New("unexpected Tx type")
	}

	return out, nil
}

func Collect[T, E any](source []T, f func(a T) E) []E {
	r := make([]E, len(source))
	for i, v := range source {
		r[i] = f(v)
	}
	return r
}
