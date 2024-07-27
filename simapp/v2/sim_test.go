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
	consensustypes "cosmossdk.io/x/consensus/types"
	"encoding/json"
	"errors"
	"fmt"
	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	cmttypes "github.com/cometbft/cometbft/types"
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
	"golang.org/x/exp/maps"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

type T = transaction.Tx
type HasWeightedOperationsX interface {
	WeightedOperationsX(weight simsx.WeightSource, reg simsx.Registry)
}

const (
	minTimePerBlock int64 = 10000 / 2

	maxTimePerBlock int64 = 10000

	timeRangePerBlock = maxTimePerBlock - minTimePerBlock
)

func TestSimsAppV2(t *testing.T) {
	DefaultNodeHome = t.TempDir()
	currentDir, err := os.Getwd()
	require.NoError(t, err)
	configPath := filepath.Join(currentDir, "testdata")
	v, err := serverv2.ReadConfig(configPath)
	require.NoError(t, err)
	v.Set("store.app-db-backend", "memdb") // todo: I had added this new type to speed up testing. Does it make sense this way?
	logger := log.NewTestLoggerInfo(t)
	app := NewSimApp[T](logger, v)

	tCfg := cli.NewConfigFromFlags().With(t, 1, nil)

	appStateFn := simtestutil.AppStateFnY(
		app.AppCodec(),
		app.AuthKeeper.AddressCodec(),
		app.StakingKeeper.ValidatorAddressCodec(),
		toSimsModule(app.ModuleManager().Modules()),
		app.DefaultGenesis(),
	)
	require.Equal(t, int64(1), tCfg.Seed) // todo revmoe
	r := rand.New(rand.NewSource(tCfg.Seed))
	params := simulation.RandomParams(r)
	accounts := slices.DeleteFunc(simtypes.RandomAccounts(r, params.NumKeys()),
		func(acc simtypes.Account) bool { // remove blocked accounts
			return app.BankKeeper.GetBlockedAddresses()[acc.AddressBech32]
		})

	appState, accounts, chainID, genesisTimestamp := appStateFn(r, accounts, tCfg)

	appStore := app.GetStore().(cometbfttypes.Store)
	//consensusParams := simulation.RandomConsensusParams(r, appState, cdc, blockMaxGas)
	req := &appmanager.BlockRequest[T]{
		Height:  1, // todo: do we start at height 1 instead of 0  in v2?
		Time:    genesisTimestamp,
		Hash:    make([]byte, 32),
		ChainId: chainID,
		AppHash: make([]byte, 32),
		ConsensusMessages: []transaction.Msg{&consensustypes.MsgUpdateParams{
			Authority: app.GetConsensusAuthority(), // todo: what else is needed?
			Block: &cmtproto.BlockParams{
				MaxBytes: 200000,
				MaxGas:   100_000_000,
			},
			Evidence: &cmtproto.EvidenceParams{
				MaxAgeNumBlocks: 302400,
				MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
				MaxBytes:        10000,
			},
			Validator: &cmtproto.ValidatorParams{PubKeyTypes: []string{cmttypes.ABCIPubKeyTypeEd25519, cmttypes.ABCIPubKeyTypeSecp256k1}},
		}},
		IsGenesis: true,
	}
	ctx, done := context.WithCancel(context.Background())
	defer done()
	_, genesisState, err := app.InitGenesis(ctx, req, appState, &genericTxDecoder[T]{txConfig: app.TxConfig()})
	require.NoError(t, err)
	changeSet, err := genesisState.GetStateChanges()
	require.NoError(t, err)
	cs := &store.Changeset{Changes: changeSet}
	stateRoot, err := appStore.Commit(cs)
	require.NoError(t, err)

	// next add a block
	emptySimParams := make(map[string]json.RawMessage, 0) // todo read sims params from disk as before
	weights := simsx.ParamWeightSource(emptySimParams)

	factoryRegistry := make(SimsV2Reg)
	// register all msg factories
	for name, v := range app.ModuleManager().Modules() {
		if name == "authz" || // todo: enable when router issue is solved with `/` prefix in MsgTypeURL
			name == "staking" { // todo: set proper consensus data and no changeset panic by x/staking/keeper/val_state_change.go:166
			continue
		}
		if w, ok := v.(HasWeightedOperationsX); ok {
			w.WeightedOperationsX(weights, factoryRegistry)
		}
	}
	// todo: register legacy and v1 msg proposals

	const ( // todo: read from CLI instead
		numBlocks     = 50 // 500 default
		maxTXPerBlock = 20 // 200 default
	)

	reporter := simsx.NewBasicSimulationReporter()
	blockTime := genesisTimestamp
	var (
		txSkippedCounter int
		txTotalCounter   int
	)
	for i := 0; i < numBlocks; i++ {
		blockTime = blockTime.Add(time.Duration(minTimePerBlock) * time.Second)
		blockTime = blockTime.Add(time.Duration(int64(r.Intn(int(timeRangePerBlock)))) * time.Second)

		reqN := &appmanager.BlockRequest[T]{
			Height:  uint64(2 + i),
			Time:    blockTime,
			Hash:    stateRoot,
			AppHash: stateRoot,
			ChainId: chainID,
		}
		cometInfo := comet.Info{
			//Evidence:        toCoreEvidence(req.Misbehavior),
			//ValidatorsHash:  req.NextValidatorsHash,
			//ProposerAddress: req.ProposerAddress,
			//LastCommit:      toCoreCommitInfo(req.DecidedLastCommit),
		}
		//app.ConsensusParamsKeeper.SetCometInfo()
		orderedFactories := maps.Values(factoryRegistry)
		slices.SortFunc(orderedFactories, func(a, b weightedFactory) int {
			switch {
			case a.weight > b.weight:
				return -1
			case a.weight < b.weight:
				return 1
			}
			return strings.Compare(sdk.MsgTypeURL(a.factory.MsgType()), sdk.MsgTypeURL(b.factory.MsgType()))
		})
		ctx = context.WithValue(ctx, corecontext.CometInfoKey, cometInfo) // required
		var txPerBlockCounter int
		blockRsp, updates, err := app.DeliverSims(ctx, reqN, func(ctx context.Context) (T, bool) {
			testData := simsx.NewChainDataSource(ctx, r, app.AuthKeeper, app.BankKeeper, app.txConfig.SigningContext().AddressCodec(), accounts...)
			for txPerBlockCounter <= maxTXPerBlock {
				txPerBlockCounter++
				// todo: sort and pick msg factory by weight
				wFac := simsx.OneOf(testData.Rand(), orderedFactories)
				sRep := reporter.WithScope(wFac.factory.MsgType())

				// the stf context is required to access state via keepers
				signers, msg := wFac.factory.Create()(ctx, testData, sRep)
				if sRep.IsSkipped() {
					txSkippedCounter++
					continue
				}
				tx, err := genTestTX(ctx, app.AuthKeeper, signers, msg, r, app.txConfig, chainID)
				require.NoError(t, err)
				return tx, false
			}
			return nil, true
		})
		require.NoError(t, err)
		changeSet, err = updates.GetStateChanges()
		require.NoError(t, err)
		stateRoot, err = appStore.Commit(&store.Changeset{Changes: changeSet})
		require.NoError(t, err)
		for _, v := range blockRsp.TxResults {
			require.NoError(t, v.Error)
		}
		txTotalCounter += txPerBlockCounter
	}
	fmt.Printf("Tx total: %d skipped: %d\n", txTotalCounter, txSkippedCounter)
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

// todo: this is the same as in commands
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
